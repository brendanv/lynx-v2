package feeds

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mmcdole/gofeed"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

// FeedResult contains the parsed feed and the ETag and Last-Modified
// headers received from the remote server
type FeedResult struct {
	Feed         *gofeed.Feed
	ETag         string
	LastModified string
}

// LoadFeedFromURL fetches and parses a feed from the given URL.
// It optionally uses etag and ifModifiedSince for conditional requests.
func LoadFeedFromURL(url string, etag string, ifModifiedSince time.Time) (*FeedResult, error) {
	fp := gofeed.NewParser()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if !ifModifiedSince.IsZero() {
		req.Header.Set("If-Modified-Since", ifModifiedSince.Format(http.TimeFormat))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return &FeedResult{}, nil
	}

	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return &FeedResult{
		Feed:         feed,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	}, nil
}

func SaveNewFeedItems(app core.App, feed *gofeed.Feed, user string, feedId string, lastArticlePubDate time.Time) error {
	collection, err := app.Dao().FindCollectionByNameOrId("feed_items")
	if err != nil {
		return err
	}
	for _, item := range feed.Items {
		if item.PublishedParsed != nil && !item.PublishedParsed.After(lastArticlePubDate) {
			continue
		}
		existingItem, _ := app.Dao().FindFirstRecordByFilter(
			"feed_items",
			"feed = {:feed} && guid = {:guid}",
			map[string]interface{}{"feed": feedId, "guid": item.GUID},
		)
		if existingItem == nil {
			newItem := models.NewRecord(collection)
			newItem.Set("user", user)
			newItem.Set("feed", feedId)
			newItem.Set("title", item.Title)
			newItem.Set("pub_date", item.PublishedParsed)
			newItem.Set("guid", item.GUID)
			newItem.Set("description", item.Description)
			newItem.Set("url", item.Link)
			if err := app.Dao().SaveRecord(newItem); err != nil {
				return err
			}
		}
	}
	return nil
}

func FetchNewFeedItems(app core.App, feedId string) error {
	feed, err := app.Dao().FindRecordById("feeds", feedId)
	if err != nil {
		return fmt.Errorf("failed to find feed: %w", err)
	}

	feedURL := feed.GetString("feed_url")
	etag := feed.GetString("etag")
	lastModified := feed.GetString("modified")
	lastModifiedTime, _ := time.Parse(http.TimeFormat, lastModified)

	feedResult, err := LoadFeedFromURL(feedURL, etag, lastModifiedTime)
	if err != nil {
		return fmt.Errorf("failed to load feed from URL: %w", err)
	}

	if feedResult.Feed == nil {
		return nil
	}

	feed.Set("etag", feedResult.ETag)
	feed.Set("modified", feedResult.LastModified)
	previousFetchTime := feed.GetDateTime("last_fetched_at").Time()
	lastFetchedAt := time.Now().UTC()
	feed.Set("last_fetched_at", lastFetchedAt.Format(time.RFC3339))
	if err := app.Dao().SaveRecord(feed); err != nil {
		return fmt.Errorf("failed to update feed record: %w", err)
	}

	if err := SaveNewFeedItems(app, feedResult.Feed, feed.GetString("user"), feedId, previousFetchTime); err != nil {
		return fmt.Errorf("failed to save new feed items: %w", err)
	}

	return nil
}

// SaveNewFeed extracts the URL from the request, loads the
// feed, and saves it to the database along with the first
// set of feed items.
func SaveNewFeed(app core.App, c echo.Context) error {
	authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	if authRecord == nil {
		return apis.NewForbiddenError("Not authenticated", nil)
	}

	url := c.FormValue("url")
	if url == "" {
		return apis.NewBadRequestError("URL is required", nil)
	}

	feedResult, err := LoadFeedFromURL(url, "", time.Time{})
	if err != nil {
		return apis.NewBadRequestError("Error parsing feed", err)
	}

	collection, err := app.Dao().FindCollectionByNameOrId("feeds")
	if err != nil {
		return apis.NewBadRequestError("Failed to find feeds collection", err)
	}

	record := models.NewRecord(collection)
	record.Set("user", authRecord.Id)
	record.Set("feed_url", url)
	record.Set("name", feedResult.Feed.Title)
	record.Set("description", feedResult.Feed.Description)
	if feedResult.Feed.Image != nil {
		record.Set("image_url", feedResult.Feed.Image.URL)
	}
	record.Set("etag", feedResult.ETag)
	record.Set("modified", feedResult.LastModified)
	record.Set("last_fetched_at", time.Now().UTC().Format(time.RFC3339))
	record.Set("auto_add_feed_items_to_library", false)

	if err := app.Dao().SaveRecord(record); err != nil {
		return apis.NewBadRequestError("Failed to save feed", err)
	}

	if err := SaveNewFeedItems(app, feedResult.Feed, authRecord.Id, record.Id, time.Time{}); err != nil {
		return apis.NewBadRequestError("Failed to save feed items", err)
	}

	return c.JSON(200, map[string]interface{}{
		"id": record.Id,
	})
}

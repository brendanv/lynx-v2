package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"github.com/go-shiori/go-readability"

	_ "main/migrations"
)

func main() {
	app := pocketbase.New()

	// Enable the migration command, but only enable
	// Automigrate in dev (if go-run is used)
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/lynx/parse_link",
			Handler: func(c echo.Context) error {
				return handleParseURL(app, c)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth(),
			},
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// Given a URL, load the URL (using relevant cookies for the authenticated
// user), extract the article content, and create a new Link record in pocketbase.
func handleParseURL(app *pocketbase.PocketBase, c echo.Context) error {
	urlParam := c.FormValue("url")
	if urlParam == "" {
		return apis.NewBadRequestError("Missing 'url' parameter", nil)
	}
	
	parsedURL, err := url.Parse(urlParam)
	if err != nil {
		return apis.NewBadRequestError("Failed to parse URL", err)
	}

	// Get the authenticated user
	authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	if authRecord == nil {
		return apis.NewForbiddenError("Not authenticated", nil)
	}

	// Load user cookies 
	cookieRecords, err := app.Dao().FindRecordsByFilter(
		"user_cookies",
		"user = {:user} && domain = {:url}",
		"-created",
		10,
		0,
		dbx.Params{"user": authRecord.Id, "url": parsedURL.Hostname()},
	)
	if err != nil {
		return apis.NewBadRequestError("Failed to fetch cookies", err)
	}

	// Prepare cookies for the request
	var cookies []*http.Cookie
	for _, record := range cookieRecords {
		cookies = append(cookies, &http.Cookie{
			Name:  record.GetString("name"),
			Value: record.GetString("value"),
		})
	}

	// Build request & add cookies
	client := &http.Client{}
	req, err := http.NewRequest("GET", urlParam, nil)
	if err != nil {
		return apis.NewBadRequestError("Failed to create request", err)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// Send
	resp, err := client.Do(req)
	if err != nil {
		return apis.NewBadRequestError("Failed to send request", err)
	}
	defer resp.Body.Close()

	// resp.Body can only be read once, so store it locally here.
	bodyContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return apis.NewBadRequestError("Failed to read response body", err)
	}

	// Use go-readability to parse the webpage
	bodyReader := strings.NewReader(string(bodyContent))
	article, err := readability.FromReader(bodyReader, parsedURL)
	if err != nil {
		return apis.NewBadRequestError("Failed to parse webpage content", err)
	}

	// Create a new record in the links collection
	collection, err := app.Dao().FindCollectionByNameOrId("links")
	if err != nil {
		return apis.NewBadRequestError("Failed to find links collection", err)
	}

	record := models.NewRecord(collection)
	record.Set("added_to_library", time.Now().Format(time.RFC3339))
	record.Set("original_url", urlParam)
	record.Set("cleaned_url", resp.Request.URL.String())
	record.Set("title", article.Title)
	record.Set("hostname", resp.Request.URL.Hostname())
	record.Set("user", authRecord.Id)
	record.Set("excerpt", article.Excerpt)
	record.Set("author", article.Byline)
	record.Set("article_html", article.Content)
	record.Set("raw_text_content", article.TextContent)
	record.Set("header_image_url", article.Image)
	record.Set("article_date", article.PublishedTime)
	record.Set("full_page_html", string(bodyContent))

	// Calculate read time, using 285 wpm as read rate
	words := strings.Fields(article.TextContent)
	wordCount := len(words)
	minutes := float64(wordCount) / float64(285)
	readTime := time.Duration(minutes * float64(time.Minute))
	record.Set("read_time_seconds", int(math.Round(readTime.Seconds())))
	record.Set("read_time_display", fmt.Sprintf("%d min", int(math.Round(readTime.Minutes()))))

	if err := app.Dao().SaveRecord(record); err != nil {
		return apis.NewBadRequestError("Failed to save link", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id": record.Id,
	})
}

package lynx

import (
	"log"
	"net/http"
	"os"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/pocketbase/pocketbase/tools/routine"
	"github.com/pocketbase/pocketbase/tools/security"

	"main/lynx/feeds"
	"main/lynx/singlefile"
	"main/lynx/summarizer"
	"main/lynx/url_parser"
)

var parseUrlHandlerFunc = url_parser.HandleParseURLRequest
var parseFeedHandlerFunc = feeds.SaveNewFeed
var convertFeedItemToLinkFunc = feeds.MaybeConvertFeedItemToLink

// Interfaces for dependency injection for summarization tests
type Summarizer interface {
	MaybeSummarizeLink(app core.App, linkID string)
}

var CurrentSummarizer Summarizer = &DefaultSummarizer{}

type DefaultSummarizer struct{}

func (s *DefaultSummarizer) MaybeSummarizeLink(app core.App, linkID string) {
	summarizer.MaybeSummarizeLink(app, linkID)
}

func InitializePocketbase(app core.App) {

	apiKeyAuth := ApiKeyAuthMiddleware(app)

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		scheduler := cron.New()

		e.Router.GET(
			"/*",
			apis.StaticDirectoryHandler(os.DirFS("./pb_public"), true),
		)

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/lynx/parse_link",
			Handler: func(c echo.Context) error {
				record, err := parseUrlHandlerFunc(app, c)
				if err != nil {
					return err
				}
				return c.JSON(http.StatusOK, map[string]interface{}{
					"id": record.Id,
				})
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				apiKeyAuth,
				apis.RequireAdminOrRecordAuth(),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/lynx/generate_api_key",
			Handler: func(c echo.Context) error {
				return handleGenerateAPIKey(app, c)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				apis.RequireAdminOrRecordAuth(),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/lynx/parse_feed",
			Handler: func(c echo.Context) error {
				return parseFeedHandlerFunc(app, c)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				apiKeyAuth,
				apis.RequireAdminOrRecordAuth(),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/lynx/link/:id/create_archive",
			Handler: func(c echo.Context) error {
				return handleArchiveLink(app, c)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				apis.RequireAdminOrRecordAuth(),
			},
		})

		scheduler.MustAdd("FetchFeeds", "0 */6 * * *", func() {
			feeds.FetchAllFeeds(app)
		})

		scheduler.Start()

		return nil
	})

	// Automatically update last_viewed_at when links are loaded
	// individually. However, let the client control this behavior
	// with a header.
	app.OnRecordViewRequest("links").Add(func(e *core.RecordViewEvent) error {
		updateHeader := e.HttpContext.Request().Header.Get("X-Lynx-Update-Last-Viewed")
		if updateHeader != "true" {
			return nil
		}

		e.Record.Set("last_viewed_at", time.Now().UTC().Format(time.RFC3339))

		err := app.Dao().SaveRecord(e.Record)
		if err != nil {
			log.Printf("Failed to update last_viewed_at: %v", err)
			return err
		}

		return nil
	})

	app.OnModelAfterCreate("links").Add(func(e *core.ModelEvent) error {
		routine.FireAndForget(func() {
			CurrentSummarizer.MaybeSummarizeLink(app, e.Model.GetId())
		})
		routine.FireAndForget(func() {
			singlefile.MaybeArchiveLink(app, e.Model.GetId())
		})
		return nil
	})

	app.OnModelAfterCreate("feed_items").Add(func(e *core.ModelEvent) error {
		routine.FireAndForget(func() {
			convertFeedItemToLinkFunc(app, e.Model.GetId())
		})
		return nil
	})

	app.OnRecordViewRequest("links").Add(func(e *core.RecordViewEvent) error {
		e.Record.WithUnknownData(true)
		converter := md.NewConverter("", true, nil)
		markdown, err := converter.ConvertString(e.Record.GetString("article_html"))
		if err != nil {
			log.Fatal(err)
		} else {
			e.Record.Set("article_markdown", markdown)
		}
		return nil
	})
}

func handleGenerateAPIKey(app core.App, c echo.Context) error {
	authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	if authRecord == nil {
		return apis.NewForbiddenError("Not authenticated", nil)
	}

	name := c.FormValue("name")
	if name == "" {
		return apis.NewBadRequestError("'name' parameter is required", nil)
	}

	apiKey := security.RandomString(32)

	// Set expiration date to 6 months from now
	expiresAt := time.Now().UTC().AddDate(0, 6, 0)

	collection, err := app.Dao().FindCollectionByNameOrId("api_keys")
	if err != nil {
		return apis.NewBadRequestError("Failed to find api_keys collection", err)
	}

	record := models.NewRecord(collection)
	record.Set("user", authRecord.Id)
	record.Set("api_key", apiKey)
	record.Set("name", name)
	record.Set("expires_at", expiresAt.Format(time.RFC3339))

	if err := app.Dao().SaveRecord(record); err != nil {
		return apis.NewBadRequestError("Failed to save API key", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"name":       name,
		"api_key":    apiKey,
		"expires_at": expiresAt.Format(time.RFC3339),
		"id":         record.Id,
	})
}

func handleArchiveLink(app core.App, c echo.Context) error {
	linkID := c.PathParam("id")
	if linkID == "" {
		return apis.NewNotFoundError("Link ID is required", nil)
	}

	authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	if authRecord == nil {
		return apis.NewForbiddenError("Not authenticated", nil)
	}

	link, err := app.Dao().FindRecordById("links", linkID)
	if err != nil {
		return apis.NewNotFoundError("Link not found", err)
	}

	if link.GetString("user") != authRecord.Id {
		return apis.NewForbiddenError("You don't have permission to archive this link", nil)
	}

	go func() {
		singlefile.MaybeArchiveLink(app, linkID)
	}()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Archive process started",
	})
}

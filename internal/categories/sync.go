package categories

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const (
	coremanagerSourceName     = "coremanager"
	coremanagerSourceCategory = 4
	coremanagerCategoryURL    = "https://pointbreak.coremanager.info/shop/index/index/categorie/4" // Categorie 4 is Oana's category
	coremanagerBaseURL        = "https://pointbreak.coremanager.info"
	categoriesCollectionName  = "categories"
)

var ticketTemplateImageIDPattern = regexp.MustCompile(`/images/tickettemplates/([0-9]+)/`)

type scrapedCategory struct {
	ExternalID  string
	Title       string
	Description string
	DetailURL   string
	ImageURL    string
	Slug        string
	Position    int
}

type syncStats struct {
	Scraped     int
	Created     int
	Updated     int
	SoftDeleted int
	Restored    int
}

type Syncer struct {
	mu sync.Mutex
}

func NewSyncer() *Syncer {
	return &Syncer{}
}

func (s *Syncer) Run(app core.App, trigger string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	started := time.Now()
	stats := syncStats{}
	app.Logger().Info("categories sync started",
		"source", coremanagerSourceName,
		"sourceCategoryID", coremanagerSourceCategory,
		"trigger", trigger,
	)

	collection, err := app.FindCollectionByNameOrId(categoriesCollectionName)
	if err != nil {
		return fmt.Errorf("find %s collection: %w", categoriesCollectionName, err)
	}
	syncTimestamp := time.Now().UTC().Format(time.RFC3339)

	categories, err := scrapeCoremanagerCategories(coremanagerCategoryURL, coremanagerBaseURL)
	if err != nil {
		return err
	}
	stats.Scraped = len(categories)

	seen := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		externalKey := buildExternalKey(coremanagerSourceName, coremanagerSourceCategory, category.ExternalID)
		seen[externalKey] = struct{}{}

		record, err := app.FindFirstRecordByData(collection, "external_key", externalKey)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("find category by external key %q: %w", externalKey, err)
			}
			record = core.NewRecord(collection)
			record.Set("external_key", externalKey)
			stats.Created++
		} else {
			stats.Updated++
		}

		record.Set("source", coremanagerSourceName)
		record.Set("source_category_id", coremanagerSourceCategory)
		record.Set("external_id", category.ExternalID)
		record.Set("slug", category.Slug)
		record.Set("title", category.Title)
		record.Set("description", category.Description)
		record.Set("detail_url", category.DetailURL)
		record.Set("image_url", category.ImageURL)
		record.Set("position", category.Position)
		record.Set("last_synced_at", syncTimestamp)
		if record.GetBool("is_deleted") {
			stats.Restored++
		}
		record.Set("is_deleted", false)

		if err := app.Save(record); err != nil {
			return fmt.Errorf("save category %q: %w", category.Title, err)
		}
	}

	existing, err := app.FindAllRecords(collection, dbx.HashExp{
		"source":             coremanagerSourceName,
		"source_category_id": coremanagerSourceCategory,
	})
	if err != nil {
		return fmt.Errorf("load existing categories: %w", err)
	}

	for _, record := range existing {
		externalKey := record.GetString("external_key")
		if _, ok := seen[externalKey]; ok {
			continue
		}

		if record.GetBool("is_deleted") {
			continue
		}

		record.Set("is_deleted", true)
		record.Set("last_synced_at", syncTimestamp)
		if err := app.Save(record); err != nil {
			return fmt.Errorf("soft delete stale category %q: %w", externalKey, err)
		}
		stats.SoftDeleted++
	}

	app.Logger().Info("categories sync finished",
		"source", coremanagerSourceName,
		"sourceCategoryID", coremanagerSourceCategory,
		"trigger", trigger,
		"durationMs", time.Since(started).Milliseconds(),
		"scraped", stats.Scraped,
		"created", stats.Created,
		"updated", stats.Updated,
		"softDeleted", stats.SoftDeleted,
		"restored", stats.Restored,
	)

	return nil
}

func scrapeCoremanagerCategories(pageURL, baseURL string) ([]scrapedCategory, error) {
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "oana-pocketbase-sync/1.0")
	req.Header.Set("Accept", "text/html")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch categories page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch categories page: status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse categories html: %w", err)
	}

	categories := make([]scrapedCategory, 0, 16)
	doc.Find("#categories .category_entry").Each(func(i int, selection *goquery.Selection) {
		imageRef := strings.TrimSpace(selection.Find("img.tile-image").First().AttrOr("src", ""))
		externalID := extractCategoryExternalID(imageRef)
		if externalID == "" {
			return
		}

		title := strings.TrimSpace(selection.Find(".category_heading").First().Text())
		description := strings.TrimSpace(selection.Find(".category_short_text").First().Text())
		detailURL := absoluteURL(baseURL, strings.TrimSpace(selection.Find("a[href]").First().AttrOr("href", "")))
		imageURL := absoluteURL(baseURL, imageRef)

		categories = append(categories, scrapedCategory{
			ExternalID:  externalID,
			Title:       title,
			Description: description,
			DetailURL:   detailURL,
			ImageURL:    imageURL,
			Slug:        strings.TrimSpace(selection.AttrOr("data-link", "")),
			Position:    i + 1,
		})
	})

	if len(categories) == 0 {
		return nil, fmt.Errorf("no categories found in source html")
	}

	return categories, nil
}

func absoluteURL(baseURL, ref string) string {
	if ref == "" {
		return ""
	}

	parsedRef, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	if parsedRef.IsAbs() {
		return parsedRef.String()
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return ref
	}

	return base.ResolveReference(parsedRef).String()
}

func buildExternalKey(source string, sourceCategoryID int, externalID string) string {
	return source + ":" + strconv.Itoa(sourceCategoryID) + ":" + externalID
}

func extractCategoryExternalID(imageRef string) string {
	matches := ticketTemplateImageIDPattern.FindStringSubmatch(imageRef)
	if len(matches) == 2 {
		return matches[1]
	}

	return ""
}

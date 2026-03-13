package sessions

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const (
	coremanagerSourceName      = "coremanager"
	coremanagerSourceCategory  = 4
	coremanagerEventboxURL     = "https://pointbreak.coremanager.info/shop/index/geteventbox"
	categoriesCollectionName   = "categories"
	sessionsCollectionName     = "sessions"
	sessionsDaysAhead          = 40
	eventboxRequestConcurrency = 8
)

type Syncer struct {
	mu sync.Mutex
}

type syncStats struct {
	Categories     int
	Requests       int
	FailedRequests int
	FetchedEvents  int
	SkippedEvents  int
	Created        int
	Updated        int
	SoftDeleted    int
	Restored       int
}

type eventboxResponse struct {
	CleanEvent []map[string]any `json:"clean_event"`
}

type eventFetchJob struct {
	CategoryID string
	DateUnix   int64
}

type sessionPayload struct {
	ExternalKey        string
	CategoryExternalID string
	RequestDateUnix    int64
	ExternalID         string
	Title              string
	Start              string
	End                string
	Location           string
	Trainer            string
	ParticipantsCount  int
	MaxParticipants    int
	RawJSON            string
}

func NewSyncer() *Syncer {
	return &Syncer{}
}

func (s *Syncer) Run(app core.App, trigger string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	started := time.Now()
	stats := syncStats{}
	app.Logger().Info("sessions sync started",
		"source", coremanagerSourceName,
		"sourceCategoryID", coremanagerSourceCategory,
		"trigger", trigger,
		"daysAhead", sessionsDaysAhead,
	)

	sessionsCollection, err := app.FindCollectionByNameOrId(sessionsCollectionName)
	if err != nil {
		return fmt.Errorf("find %s collection: %w", sessionsCollectionName, err)
	}
	participantsCollection, err := app.FindCollectionByNameOrId(sessionParticipantsCollectionName)
	if err != nil {
		return fmt.Errorf("find %s collection: %w", sessionParticipantsCollectionName, err)
	}
	syncTimestamp := time.Now().UTC().Format(time.RFC3339)

	categoryIDs, err := loadActiveCategoryIDs(app)
	if err != nil {
		return err
	}
	stats.Categories = len(categoryIDs)

	jobs := make([]eventFetchJob, 0, len(categoryIDs)*(sessionsDaysAhead+1))
	currentDateUnix := time.Now().Unix()
	for day := 0; day <= sessionsDaysAhead; day++ {
		dateUnix := currentDateUnix + int64(day*24*60*60)
		for _, categoryID := range categoryIDs {
			jobs = append(jobs, eventFetchJob{
				CategoryID: categoryID,
				DateUnix:   dateUnix,
			})
		}
	}
	stats.Requests = len(jobs)

	seen := make(map[string]sessionPayload, 1024)
	var seenMu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, eventboxRequestConcurrency)

	for _, job := range jobs {
		wg.Add(1)
		go func(job eventFetchJob) {
			defer wg.Done()

			sem <- struct{}{}
			events, fetchErr := fetchCoremanagerEvents(job.CategoryID, job.DateUnix)
			<-sem

			if fetchErr != nil {
				app.Logger().Warn("sessions sync request failed",
					"categoryExternalID", job.CategoryID,
					"dateUnix", job.DateUnix,
					"error", fetchErr,
				)
				seenMu.Lock()
				stats.FailedRequests++
				seenMu.Unlock()
				return
			}

			seenMu.Lock()
			stats.FetchedEvents += len(events)
			seenMu.Unlock()

			for _, event := range events {
				payload, ok := buildSessionPayload(job.CategoryID, job.DateUnix, event)
				seenMu.Lock()
				if !ok {
					stats.SkippedEvents++
					seenMu.Unlock()
					continue
				}
				seen[payload.ExternalKey] = payload
				seenMu.Unlock()
			}
		}(job)
	}

	wg.Wait()

	for _, payload := range seen {
		record, err := app.FindFirstRecordByData(sessionsCollection, "external_key", payload.ExternalKey)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("find session by external key %q: %w", payload.ExternalKey, err)
			}
			record = core.NewRecord(sessionsCollection)
			record.Set("external_key", payload.ExternalKey)
			stats.Created++
		} else {
			stats.Updated++
		}

		record.Set("source", coremanagerSourceName)
		record.Set("source_category_id", coremanagerSourceCategory)
		record.Set("category_external_id", payload.CategoryExternalID)
		record.Set("request_date_unix", payload.RequestDateUnix)
		record.Set("external_id", payload.ExternalID)
		record.Set("title", payload.Title)
		record.Set("start", payload.Start)
		record.Set("end", payload.End)
		record.Set("location", payload.Location)
		record.Set("trainer", payload.Trainer)
		record.Set("participants_count", payload.ParticipantsCount)
		record.Set("participants_synced_at", syncTimestamp)
		record.Set("max_participants", payload.MaxParticipants)
		features := deriveTemporalFeatures(payload.Start, payload.RequestDateUnix)
		record.Set("event_date", features.EventDate)
		record.Set("event_weekday", features.EventWeekday)
		record.Set("event_start_time", features.EventStartTime)
		record.Set("event_start_hour", features.EventStartHour)
		record.Set("event_start_minute", features.EventStartMinute)
		record.Set("event_time_known", features.EventTimeKnown)
		record.Set("last_synced_at", syncTimestamp)
		record.Set("raw", payload.RawJSON)
		if record.GetBool("is_deleted") {
			stats.Restored++
		}
		record.Set("is_deleted", false)

		if err := app.Save(record); err != nil {
			return fmt.Errorf("save session %q: %w", payload.ExternalKey, err)
		}

		if _, err := appendSnapshotIfCountChanged(
			app,
			participantsCollection,
			record.Id,
			payload.ParticipantsCount,
			syncTimestamp,
			payload.Start,
		); err != nil {
			return err
		}
	}

	existing, err := app.FindAllRecords(sessionsCollection, dbx.HashExp{
		"source":             coremanagerSourceName,
		"source_category_id": coremanagerSourceCategory,
	})
	if err != nil {
		return fmt.Errorf("load existing sessions: %w", err)
	}

	loc := loadAnalyticsLocation()
	now := time.Now().In(loc)

	for _, record := range existing {
		externalKey := record.GetString("external_key")
		if _, ok := seen[externalKey]; ok {
			continue
		}

		if record.GetBool("is_deleted") {
			continue
		}

		if sessionHasStarted(record, now, loc) {
			continue
		}

		record.Set("is_deleted", true)
		record.Set("last_synced_at", syncTimestamp)
		if err := app.Save(record); err != nil {
			return fmt.Errorf("soft delete stale session %q: %w", externalKey, err)
		}
		stats.SoftDeleted++
	}

	app.Logger().Info("sessions sync finished",
		"source", coremanagerSourceName,
		"sourceCategoryID", coremanagerSourceCategory,
		"trigger", trigger,
		"durationMs", time.Since(started).Milliseconds(),
		"categories", stats.Categories,
		"requests", stats.Requests,
		"failedRequests", stats.FailedRequests,
		"fetchedEvents", stats.FetchedEvents,
		"skippedEvents", stats.SkippedEvents,
		"created", stats.Created,
		"updated", stats.Updated,
		"softDeleted", stats.SoftDeleted,
		"restored", stats.Restored,
	)

	return nil
}

func sessionHasStarted(record *core.Record, now time.Time, loc *time.Location) bool {
	if start, ok := parseEventStart(record.GetString("start"), loc); ok {
		return !start.After(now)
	}

	eventDate := strings.TrimSpace(record.GetString("event_date"))
	if eventDate == "" {
		return false
	}

	parsedDate, err := time.ParseInLocation("2006-01-02", eventDate, loc)
	if err != nil {
		return false
	}

	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return parsedDate.Before(startOfToday)
}

func loadActiveCategoryIDs(app core.App) ([]string, error) {
	categoriesCollection, err := app.FindCollectionByNameOrId(categoriesCollectionName)
	if err != nil {
		return nil, fmt.Errorf("find %s collection: %w", categoriesCollectionName, err)
	}

	categories, err := app.FindAllRecords(categoriesCollection, dbx.HashExp{
		"source":             coremanagerSourceName,
		"source_category_id": coremanagerSourceCategory,
		"is_deleted":         false,
	})
	if err != nil {
		return nil, fmt.Errorf("load active categories: %w", err)
	}

	idSet := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		id := strings.TrimSpace(category.GetString("external_id"))
		if id == "" {
			continue
		}
		idSet[id] = struct{}{}
	}

	categoryIDs := make([]string, 0, len(idSet))
	for id := range idSet {
		categoryIDs = append(categoryIDs, id)
	}
	sort.Strings(categoryIDs)

	if len(categoryIDs) == 0 {
		return nil, fmt.Errorf("no active category external_ids found")
	}

	return categoryIDs, nil
}

func fetchCoremanagerEvents(categoryID string, dateUnix int64) ([]map[string]any, error) {
	requestURL, err := buildEventboxURL(categoryID, dateUnix)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "oana-pocketbase-sync/1.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request eventbox: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request eventbox: status %d", resp.StatusCode)
	}

	var parsed eventboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		// Coremanager can return non-json for empty days.
		return nil, nil
	}

	if len(parsed.CleanEvent) == 0 {
		return nil, nil
	}

	return parsed.CleanEvent, nil
}

func buildEventboxURL(categoryID string, dateUnix int64) (string, error) {
	parsedURL, err := url.Parse(coremanagerEventboxURL)
	if err != nil {
		return "", fmt.Errorf("parse eventbox url: %w", err)
	}

	query := parsedURL.Query()
	query.Set("id", categoryID)
	query.Set("date", strconv.FormatInt(dateUnix, 10))
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

func buildSessionPayload(categoryID string, dateUnix int64, event map[string]any) (sessionPayload, bool) {
	rawJSON, err := json.Marshal(event)
	if err != nil {
		return sessionPayload{}, false
	}

	externalID := firstNonEmpty(event,
		"id",
		"event_id",
		"eventid",
		"eventId",
		"uuid",
	)
	title := firstNonEmpty(event, "title", "name", "event_name")
	start := firstNonEmpty(event, "start", "start_time", "startTime", "date_start")
	end := firstNonEmpty(event, "end", "end_time", "endTime", "date_end")
	location := firstNonEmpty(event, "location", "room", "studio")
	trainer := firstNonEmpty(event, "trainer", "teacher", "coach", "instructor")
	participantsCount, _ := extractParticipantsCount(event)
	maxParticipants, _ := extractMaxParticipants(event)

	externalKey := buildSessionExternalKey(categoryID, externalID, start, end, title)
	if externalKey == "" {
		return sessionPayload{}, false
	}

	return sessionPayload{
		ExternalKey:        externalKey,
		CategoryExternalID: categoryID,
		RequestDateUnix:    dateUnix,
		ExternalID:         externalID,
		Title:              title,
		Start:              start,
		End:                end,
		Location:           location,
		Trainer:            trainer,
		ParticipantsCount:  participantsCount,
		MaxParticipants:    maxParticipants,
		RawJSON:            string(rawJSON),
	}, true
}

func buildSessionExternalKey(categoryID, externalID, start, end, title string) string {
	if externalID != "" {
		return coremanagerSourceName + ":" + strconv.Itoa(coremanagerSourceCategory) + ":" + categoryID + ":" + externalID
	}

	parts := []string{
		coremanagerSourceName,
		strconv.Itoa(coremanagerSourceCategory),
		categoryID,
		strings.TrimSpace(start),
		strings.TrimSpace(end),
		strings.TrimSpace(title),
	}
	joined := strings.Join(parts, "|")
	if strings.Trim(joined, "|") == "" {
		return ""
	}

	sum := sha1.Sum([]byte(joined))
	return coremanagerSourceName + ":" + strconv.Itoa(coremanagerSourceCategory) + ":" + categoryID + ":h:" + hex.EncodeToString(sum[:])
}

func firstNonEmpty(event map[string]any, keys ...string) string {
	for _, key := range keys {
		raw, ok := event[key]
		if !ok || raw == nil {
			continue
		}

		value := strings.TrimSpace(anyToString(raw))
		if value != "" {
			return value
		}
	}

	return ""
}

func anyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

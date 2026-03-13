package sessions

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const sessionParticipantsCollectionName = "session_participants"

type ParticipantsSyncer struct {
	mu sync.Mutex
}

type participantsSyncStats struct {
	Categories       int
	Requests         int
	FailedRequests   int
	FetchedEvents    int
	SnapshotsCreated int
	Unchanged        int
}

type participantPayload struct {
	SessionKey         string
	CategoryExternalID string
	RequestDateUnix    int64
	SessionExternalID  string
	SessionTitle       string
	SessionStart       string
	SessionEnd         string
	ParticipantsCount  int
	MaxParticipants    int
}

func NewParticipantsSyncer() *ParticipantsSyncer {
	return &ParticipantsSyncer{}
}

func (s *ParticipantsSyncer) RunTodayTomorrow(app core.App, trigger string) error {
	return s.runRange(app, trigger, 0, 1)
}

func (s *ParticipantsSyncer) RunFuture(app core.App, trigger string) error {
	return s.runRange(app, trigger, 2, sessionsDaysAhead)
}

func (s *ParticipantsSyncer) runRange(app core.App, trigger string, startDayOffset int, endDayOffset int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if endDayOffset < startDayOffset {
		return nil
	}

	started := time.Now()
	stats := participantsSyncStats{}
	app.Logger().Info("session participants sync started",
		"source", coremanagerSourceName,
		"sourceCategoryID", coremanagerSourceCategory,
		"trigger", trigger,
		"startDayOffset", startDayOffset,
		"endDayOffset", endDayOffset,
	)

	participantsCollection, err := app.FindCollectionByNameOrId(sessionParticipantsCollectionName)
	if err != nil {
		return fmt.Errorf("find %s collection: %w", sessionParticipantsCollectionName, err)
	}
	sessionsCollection, err := app.FindCollectionByNameOrId(sessionsCollectionName)
	if err != nil {
		return fmt.Errorf("find %s collection: %w", sessionsCollectionName, err)
	}
	syncTimestamp := time.Now().UTC().Format(time.RFC3339)

	categoryIDs, err := loadActiveCategoryIDs(app)
	if err != nil {
		return err
	}
	stats.Categories = len(categoryIDs)

	currentDateUnix := time.Now().Unix()
	targetDays := make([]int64, 0, endDayOffset-startDayOffset+1)
	for day := startDayOffset; day <= endDayOffset; day++ {
		dateUnix := currentDateUnix + int64(day*24*60*60)
		targetDays = append(targetDays, dateUnix)
	}

	jobs := make([]eventFetchJob, 0, len(categoryIDs)*len(targetDays))
	for _, dayUnix := range targetDays {
		for _, categoryID := range categoryIDs {
			jobs = append(jobs, eventFetchJob{
				CategoryID: categoryID,
				DateUnix:   dayUnix,
			})
		}
	}
	stats.Requests = len(jobs)

	seen := make(map[string]participantPayload, 1024)
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
				app.Logger().Warn("session participants sync request failed",
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
				payload, ok := buildParticipantPayload(job.CategoryID, job.DateUnix, event)
				if !ok {
					continue
				}

				seenMu.Lock()
				seen[payload.SessionKey] = payload
				seenMu.Unlock()
			}
		}(job)
	}

	wg.Wait()

	for _, payload := range seen {
		sessionRecord, err := upsertSessionParticipantFields(app, sessionsCollection, payload, syncTimestamp)
		if err != nil {
			return err
		}

		created, err := appendSnapshotIfCountChanged(
			app,
			participantsCollection,
			sessionRecord.Id,
			payload.ParticipantsCount,
			syncTimestamp,
			payload.SessionStart,
		)
		if err != nil {
			return err
		}
		if created {
			stats.SnapshotsCreated++
		} else {
			stats.Unchanged++
		}
	}

	app.Logger().Info("session participants sync finished",
		"source", coremanagerSourceName,
		"sourceCategoryID", coremanagerSourceCategory,
		"trigger", trigger,
		"startDayOffset", startDayOffset,
		"endDayOffset", endDayOffset,
		"durationMs", time.Since(started).Milliseconds(),
		"categories", stats.Categories,
		"requests", stats.Requests,
		"failedRequests", stats.FailedRequests,
		"fetchedEvents", stats.FetchedEvents,
		"snapshotsCreated", stats.SnapshotsCreated,
		"unchanged", stats.Unchanged,
	)

	return nil
}

func buildParticipantPayload(categoryID string, dateUnix int64, event map[string]any) (participantPayload, bool) {
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
	sessionKey := buildSessionExternalKey(categoryID, externalID, start, end, title)
	if sessionKey == "" {
		return participantPayload{}, false
	}

	participantsCount, _ := extractParticipantsCount(event)
	maxParticipants, _ := extractMaxParticipants(event)

	return participantPayload{
		SessionKey:         sessionKey,
		CategoryExternalID: categoryID,
		RequestDateUnix:    dateUnix,
		SessionExternalID:  externalID,
		SessionTitle:       title,
		SessionStart:       start,
		SessionEnd:         end,
		ParticipantsCount:  participantsCount,
		MaxParticipants:    maxParticipants,
	}, true
}

func upsertSessionParticipantFields(
	app core.App,
	sessionsCollection *core.Collection,
	payload participantPayload,
	syncTimestamp string,
) (*core.Record, error) {
	record, err := app.FindFirstRecordByData(sessionsCollection, "external_key", payload.SessionKey)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("find session by external key %q: %w", payload.SessionKey, err)
		}

		record = core.NewRecord(sessionsCollection)
		record.Set("external_key", payload.SessionKey)
		record.Set("source", coremanagerSourceName)
		record.Set("source_category_id", coremanagerSourceCategory)
		record.Set("category_external_id", payload.CategoryExternalID)
		record.Set("request_date_unix", payload.RequestDateUnix)
		record.Set("external_id", payload.SessionExternalID)
		record.Set("title", payload.SessionTitle)
		record.Set("start", payload.SessionStart)
		record.Set("end", payload.SessionEnd)
		record.Set("is_deleted", false)
	} else if record.GetBool("is_deleted") {
		record.Set("is_deleted", false)
	}

	record.Set("participants_count", payload.ParticipantsCount)
	record.Set("max_participants", payload.MaxParticipants)
	record.Set("participants_synced_at", syncTimestamp)
	features := deriveTemporalFeatures(payload.SessionStart, payload.RequestDateUnix)
	record.Set("event_date", features.EventDate)
	record.Set("event_weekday", features.EventWeekday)
	record.Set("event_start_time", features.EventStartTime)
	record.Set("event_start_hour", features.EventStartHour)
	record.Set("event_start_minute", features.EventStartMinute)
	record.Set("event_time_known", features.EventTimeKnown)
	record.Set("last_synced_at", syncTimestamp)

	if err := app.Save(record); err != nil {
		return nil, fmt.Errorf("save session participant fields %q: %w", payload.SessionKey, err)
	}

	return record, nil
}

func appendSnapshotIfCountChanged(
	app core.App,
	participantsCollection *core.Collection,
	sessionRecordID string,
	participantsCount int,
	recordedAt string,
	sessionStart string,
) (bool, error) {
	latest, err := app.FindRecordsByFilter(
		participantsCollection,
		"session = {:sessionId}",
		"-recorded_at",
		1,
		0,
		dbx.Params{"sessionId": sessionRecordID},
	)
	if err != nil {
		return false, fmt.Errorf("load latest participant snapshot for session %q: %w", sessionRecordID, err)
	}

	if len(latest) > 0 && latest[0].GetInt("participants_count") == participantsCount {
		return false, nil
	}

	record := core.NewRecord(participantsCollection)
	record.Set("session", sessionRecordID)
	record.Set("recorded_at", recordedAt)
	record.Set("participants_count", participantsCount)
	if eventStartAt, minutesBeforeStart, ok := deriveLeadTime(sessionStart, recordedAt); ok {
		record.Set("event_start_at", eventStartAt)
		record.Set("minutes_before_start", minutesBeforeStart)
	}
	if err := app.Save(record); err != nil {
		return false, fmt.Errorf("save participant snapshot for session %q: %w", sessionRecordID, err)
	}

	return true, nil
}

func deriveLeadTime(sessionStart string, recordedAt string) (string, int, bool) {
	loc := loadAnalyticsLocation()
	startAt, ok := parseEventStart(sessionStart, loc)
	if !ok {
		return "", 0, false
	}

	recorded, err := time.Parse(time.RFC3339, recordedAt)
	if err != nil {
		return "", 0, false
	}

	eventStartAt := startAt.UTC().Format(time.RFC3339)
	minutesBeforeStart := int(startAt.Sub(recorded).Minutes())
	return eventStartAt, minutesBeforeStart, true
}

func extractParticipantsCount(event map[string]any) (int, bool) {
	candidates := []string{
		"participants",
		"participantcount",
		"participantscount",
		"participant_count",
		"participants_count",
		"booked",
		"bookedcount",
		"booked_count",
		"attendees",
		"attendeecount",
		"attendee_count",
		"registered",
		"registrations",
	}

	for _, candidate := range candidates {
		if value, ok := findNumericValue(event, normalizeKey(candidate)); ok {
			return value, true
		}
	}

	return 0, false
}

func extractMaxParticipants(event map[string]any) (int, bool) {
	candidates := []string{
		"max_participants",
		"maxparticipants",
		"maxparticipantscount",
		"participant_limit",
		"participants_limit",
		"capacity",
		"max_bookings",
	}

	for _, candidate := range candidates {
		if value, ok := findNumericValue(event, normalizeKey(candidate)); ok {
			return value, true
		}
	}

	return 0, false
}

func findNumericValue(value any, wantedNormalizedKey string) (int, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			if normalizeKey(key) == wantedNormalizedKey {
				if parsed, ok := toInt(nested); ok {
					return parsed, true
				}
			}
			if parsed, ok := findNumericValue(nested, wantedNormalizedKey); ok {
				return parsed, true
			}
		}
	case []any:
		for _, nested := range typed {
			if parsed, ok := findNumericValue(nested, wantedNormalizedKey); ok {
				return parsed, true
			}
		}
	}

	return 0, false
}

func normalizeKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, "-", "")
	key = strings.ReplaceAll(key, " ", "")
	return key
}

func toInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case int32:
		return int(typed), true
	case float64:
		return int(typed), true
	case float32:
		return int(typed), true
	case json.Number:
		i, err := typed.Int64()
		if err != nil {
			f, ferr := typed.Float64()
			if ferr != nil {
				return 0, false
			}
			return int(f), true
		}
		return int(i), true
	case string:
		s := strings.TrimSpace(typed)
		if s == "" {
			return 0, false
		}
		i, err := strconv.Atoi(s)
		if err == nil {
			return i, true
		}
		f, ferr := strconv.ParseFloat(s, 64)
		if ferr != nil {
			return 0, false
		}
		return int(f), true
	default:
		return 0, false
	}
}

package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

const (
	categoriesCollectionName          = "categories"
	sessionsCollectionName            = "sessions"
	sessionParticipantsCollectionName = "session_participants"
)

func init() {
	m.Register(func(app core.App) error {
		if _, err := app.FindCollectionByNameOrId(categoriesCollectionName); err == nil {
			// continue to sessions creation for already bootstrapped local states
		} else {
			categories := core.NewBaseCollection(categoriesCollectionName)
			categories.Fields.Add(
				&core.AutodateField{Name: "created", OnCreate: true},
				&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
				&core.TextField{Name: "external_key", Required: true, Max: 255},
				&core.TextField{Name: "source", Required: true, Max: 120},
				&core.NumberField{Name: "source_category_id", Required: true, OnlyInt: true},
				&core.TextField{Name: "external_id", Required: true, Max: 255},
				&core.TextField{Name: "slug", Max: 255},
				&core.TextField{Name: "title", Required: true, Max: 255},
				&core.TextField{Name: "description", Max: 4000},
				&core.TextField{Name: "detail_url", Max: 1000},
				&core.TextField{Name: "image_url", Max: 1000},
				&core.NumberField{Name: "position", OnlyInt: true},
				&core.DateField{Name: "last_synced_at"},
				&core.BoolField{Name: "is_deleted"},
			)
			categories.AddIndex("idx_categories_external_key", true, "external_key", "")
			categories.AddIndex("idx_categories_source_category", false, "source, source_category_id", "")
			categories.AddIndex("idx_categories_last_synced_at", false, "last_synced_at", "")
			categories.AddIndex("idx_categories_is_deleted", false, "is_deleted", "")

			if err := app.Save(categories); err != nil {
				return err
			}
		}

		var sessions *core.Collection
		if existing, err := app.FindCollectionByNameOrId(sessionsCollectionName); err == nil {
			sessions = existing
			// continue to participants creation for already bootstrapped local states
		} else {
			sessions = core.NewBaseCollection(sessionsCollectionName)
			publicRule := types.Pointer("")
			sessions.ListRule = publicRule
			sessions.ViewRule = publicRule
			sessions.Fields.Add(
				&core.AutodateField{Name: "created", OnCreate: true},
				&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
				&core.TextField{Name: "external_key", Required: true, Max: 255},
				&core.TextField{Name: "source", Required: true, Max: 120},
				&core.NumberField{Name: "source_category_id", Required: true, OnlyInt: true},
				&core.TextField{Name: "category_external_id", Required: true, Max: 255},
				&core.NumberField{Name: "request_date_unix", Required: true, OnlyInt: true},
				&core.TextField{Name: "external_id", Max: 255},
				&core.TextField{Name: "title", Max: 255},
				&core.TextField{Name: "start", Max: 120},
				&core.TextField{Name: "end", Max: 120},
				&core.TextField{Name: "location", Max: 255},
				&core.TextField{Name: "trainer", Max: 255},
				&core.NumberField{Name: "participants_count", OnlyInt: true},
				&core.NumberField{Name: "max_participants", OnlyInt: true},
				&core.DateField{Name: "participants_synced_at"},
				&core.TextField{Name: "event_date", Max: 10},
				&core.NumberField{Name: "event_weekday", OnlyInt: true},
				&core.TextField{Name: "event_start_time", Max: 5},
				&core.NumberField{Name: "event_start_hour", OnlyInt: true},
				&core.NumberField{Name: "event_start_minute", OnlyInt: true},
				&core.BoolField{Name: "event_time_known"},
				&core.DateField{Name: "last_synced_at"},
				&core.JSONField{Name: "raw"},
				&core.BoolField{Name: "is_deleted"},
			)
			sessions.AddIndex("idx_sessions_external_key", true, "external_key", "")
			sessions.AddIndex("idx_sessions_source_category", false, "source, source_category_id", "")
			sessions.AddIndex("idx_sessions_start", false, "start", "")
			sessions.AddIndex("idx_sessions_event_date", false, "event_date", "")
			sessions.AddIndex("idx_sessions_event_weekday", false, "event_weekday", "")
			sessions.AddIndex("idx_sessions_event_start_hour", false, "event_start_hour", "")
			sessions.AddIndex("idx_sessions_participants_synced_at", false, "participants_synced_at", "")
			sessions.AddIndex("idx_sessions_last_synced_at", false, "last_synced_at", "")
			sessions.AddIndex("idx_sessions_is_deleted", false, "is_deleted", "")

			if err := app.Save(sessions); err != nil {
				return err
			}
		}

		if _, err := app.FindCollectionByNameOrId(sessionParticipantsCollectionName); err == nil {
			return nil
		}

		participants := core.NewBaseCollection(sessionParticipantsCollectionName)
		participants.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
			&core.RelationField{
				Name:          "session",
				CollectionId:  sessions.Id,
				CascadeDelete: true,
				Required:      true,
				MaxSelect:     1,
			},
			&core.DateField{Name: "recorded_at", Required: true},
			&core.NumberField{Name: "participants_count", OnlyInt: true},
			&core.DateField{Name: "event_start_at"},
			&core.NumberField{Name: "minutes_before_start", OnlyInt: true},
		)
		participants.AddIndex("idx_session_participants_session", false, "session", "")
		participants.AddIndex("idx_session_participants_recorded_at", false, "recorded_at", "")
		participants.AddIndex("idx_session_participants_event_start_at", false, "event_start_at", "")
		participants.AddIndex("idx_session_participants_minutes_before_start", false, "minutes_before_start", "")

		return app.Save(participants)
	}, func(app core.App) error {
		if participants, err := app.FindCollectionByNameOrId(sessionParticipantsCollectionName); err == nil {
			if err := app.Delete(participants); err != nil {
				return err
			}
		}

		if sessions, err := app.FindCollectionByNameOrId(sessionsCollectionName); err == nil {
			if err := app.Delete(sessions); err != nil {
				return err
			}
		}

		if categories, err := app.FindCollectionByNameOrId(categoriesCollectionName); err == nil {
			return app.Delete(categories)
		}

		return nil
	})
}

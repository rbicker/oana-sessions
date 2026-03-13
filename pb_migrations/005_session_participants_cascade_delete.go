package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		participants, err := app.FindCollectionByNameOrId(sessionParticipantsCollectionName)
		if err != nil {
			return err
		}

		field, ok := participants.Fields.GetByName("session").(*core.RelationField)
		if !ok || field == nil {
			return nil
		}

		field.CascadeDelete = true

		return app.Save(participants)
	}, func(app core.App) error {
		participants, err := app.FindCollectionByNameOrId(sessionParticipantsCollectionName)
		if err != nil {
			return err
		}

		field, ok := participants.Fields.GetByName("session").(*core.RelationField)
		if !ok || field == nil {
			return nil
		}

		field.CascadeDelete = false

		return app.Save(participants)
	})
}

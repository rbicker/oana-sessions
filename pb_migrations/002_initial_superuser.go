package migrations

import (
	"os"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		email := os.Getenv("SUPERUSER_EMAIL")
		password := os.Getenv("SUPERUSER_PASSWORD")

		if email == "" || password == "" {
			return nil // skip creating superuser if env vars are not set
		}

		superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
		if err != nil {
			return err
		}

		record := core.NewRecord(superusers)
		record.Set("email", email)
		record.Set("password", password)

		return app.Save(record)
	}, func(app core.App) error { // optional revert operation
		email := os.Getenv("SUPERUSER_EMAIL")
		record, _ := app.FindAuthRecordByEmail(core.CollectionNameSuperusers, email)
		if record == nil {
			return nil // probably already deleted
		}

		return app.Delete(record)
	})
}

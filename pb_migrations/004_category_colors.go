package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
	internalcategories "github.com/rbicker/oana-sessions/internal/categories"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId(categoriesCollectionName)
		if err != nil {
			return err
		}

		publicRule := types.Pointer("")
		collection.ListRule = publicRule
		collection.ViewRule = publicRule

		if collection.Fields.GetByName("color") == nil {
			collection.Fields.AddAt(3, &core.TextField{Name: "color", Required: true, Max: 32})
		}

		if err := app.Save(collection); err != nil {
			return err
		}

		records, err := app.FindAllRecords(collection)
		if err != nil {
			return err
		}

		for _, record := range records {
			record.Set("color", internalcategories.ColorForExternalID(record.GetString("external_id")))
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId(categoriesCollectionName)
		if err != nil {
			return err
		}

		collection.ListRule = nil
		collection.ViewRule = nil
		collection.Fields.RemoveByName("color")

		return app.Save(collection)
	})
}

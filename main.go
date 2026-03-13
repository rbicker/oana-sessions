package main

import (
	"log"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/rbicker/oana-sessions/internal/categories"
	"github.com/rbicker/oana-sessions/internal/sessions"
	_ "github.com/rbicker/oana-sessions/pb_migrations"
)

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Dashboard
		// (the isGoRun check is to enable it only during development)
		//Automigrate: isGoRun,
	})

	categoriesSyncer := categories.NewSyncer()
	sessionsSyncer := sessions.NewSyncer()
	participantsSyncer := sessions.NewParticipantsSyncer()

	app.OnBootstrap().BindFunc(func(be *core.BootstrapEvent) error {
		if err := be.Next(); err != nil {
			return err
		}

		if err := be.App.RunAppMigrations(); err != nil {
			return err
		}

		if err := be.App.ReloadCachedCollections(); err != nil {
			return err
		}
		return nil
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		if err := se.App.Cron().Add("sync_coremanager_categories_daily", "0 2 * * *", func() {
			if err := categoriesSyncer.Run(se.App, "daily_cron"); err != nil {
				se.App.Logger().Error("categories sync failed", "trigger", "daily_cron", "error", err)
			}
		}); err != nil {
			app.Logger().Error("failed to add daily categories sync cron job", "error", err)
		}

		if err := se.App.Cron().Add("sync_coremanager_sessions_hourly", "0 * * * *", func() {
			if err := sessionsSyncer.Run(se.App, "hourly_cron"); err != nil {
				se.App.Logger().Error("sessions sync failed", "trigger", "hourly_cron", "error", err)
			}

			if err := participantsSyncer.RunFuture(se.App, "hourly_cron"); err != nil {
				se.App.Logger().Error("session participants sync failed", "trigger", "hourly_cron", "error", err)
			}
		}); err != nil {
			app.Logger().Error("failed to add hourly sessions sync cron job", "error", err)
		}

		if err := se.App.Cron().Add("sync_coremanager_session_participants_5min", "*/5 * * * *", func() {
			if err := participantsSyncer.RunTodayTomorrow(se.App, "5min_cron"); err != nil {
				se.App.Logger().Error("session participants sync failed", "trigger", "5min_cron", "error", err)
			}
		}); err != nil {
			app.Logger().Error("failed to add 5min session participants sync cron job", "error", err)
		}

		go func() {
			if err := categoriesSyncer.Run(se.App, "startup"); err != nil {
				se.App.Logger().Error("categories sync failed", "trigger", "startup", "error", err)
			}

			if err := sessionsSyncer.Run(se.App, "startup"); err != nil {
				se.App.Logger().Error("sessions sync failed", "trigger", "startup", "error", err)
			}
		}()

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

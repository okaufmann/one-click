package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/janlauber/one-click/hooks"
	"github.com/janlauber/one-click/pkg/controller"
	"github.com/janlauber/one-click/pkg/env"
	"github.com/janlauber/one-click/pkg/k8s"
	"github.com/labstack/echo/v5"
	"github.com/madflojo/tasks"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func defaultPublicDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// most likely ran with go run
		return "./pb_public"
	}

	return filepath.Join(os.Args[0], "../pb_public")
}

func init() {
	env.Init()
	k8s.Init()
}

func main() {
	app := pocketbase.New()

	var publicDirFlag string

	// add "--publicDir" option flag
	app.RootCmd.PersistentFlags().StringVar(
		&publicDirFlag,
		"publicDir",
		defaultPublicDir(),
		"the directory to serve static files",
	)
	migrationsDir := "" // default to "pb_migrations" (for js) and "migrations" (for go)

	// load js files to allow loading external JavaScript migrations
	jsvm.MustRegister(app, jsvm.Config{
		// Dir: migrationsDir,
		MigrationsDir: migrationsDir,
	})

	// register the `migrate` command
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS, // or migratecmd.TemplateLangGo (default)
		Dir:          migrationsDir,
		Automigrate:  true,
	})

	// call this only if you want to use the configurable "hooks" functionality
	hooks.PocketBaseInit(app)

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS(publicDirFlag), true))

		return nil
	})

	app.OnRecordBeforeCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		switch e.Collection.Name {
		case "rollouts":
			return controller.HandleRolloutCreate(e, app)
		}
		return nil
	})

	app.OnRecordBeforeUpdateRequest().Add(func(e *core.RecordUpdateEvent) error {
		switch e.Collection.Name {
		case "rollouts":
			return controller.HandleRolloutUpdate(e, app)
		}
		return nil
	})

	app.OnRecordBeforeDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
		switch e.Collection.Name {
		case "rollouts":
			return controller.HandleRolloutDelete(e, app)
		case "projects":
			return controller.HandleProjectDelete(e, app)
		}
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// get status of a specific rollout
		e.Router.GET("/rollouts/:projectId/:rolloutId/status", func(c echo.Context) error {
			// TODO: Auth

			projectId := c.PathParam("projectId")
			rolloutId := c.PathParam("rolloutId")

			return controller.HandleRolloutStatus(c, app, projectId, rolloutId)
		})

		e.Router.GET("/rollouts/:projectId/:rolloutId/metrics", func(c echo.Context) error {
			// TODO: Auth

			projectId := c.PathParam("projectId")
			rolloutId := c.PathParam("rolloutId")

			return controller.HandleRolloutMetrics(c, app, projectId, rolloutId)
		})

		e.Router.GET("/rollouts/:projectId/:rolloutId/events", func(c echo.Context) error {
			// TODO: Auth

			projectId := c.PathParam("projectId")
			rolloutId := c.PathParam("rolloutId")

			return controller.HandleRolloutEvents(c, app, projectId, rolloutId)
		})

		e.Router.GET("/rollouts/:projectId/:podName/logs", func(c echo.Context) error {
			// TODO: Auth

			projectId := c.PathParam("projectId")
			podName := c.PathParam("podName")

			return k8s.GetRolloutLogs(c.Response().Writer, projectId, podName)
		})

		e.Router.POST("/auto-update/:autoUpdateId", func(c echo.Context) error {
			// TODO: Auth

			autoUpdateId := c.PathParam("autoUpdateId")

			return controller.HandleAutoUpdate(c, app, autoUpdateId)
		})

		return nil
	})

	scheduler := tasks.New()
	defer scheduler.Stop()

	// Add a task
	_, err := scheduler.Add(&tasks.Task{
		Interval: 1 * time.Minute,
		TaskFunc: func() error {

			err := controller.AutoUpdateController(app)

			return err
		},
	})
	if err != nil {
		log.Println(err)
	}

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

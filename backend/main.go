package main

import (
	"log"
	"os"
	"strings"
	"pocketvue/config"
	"pocketvue/hooks"
	"pocketvue/routes"
	"pocketvue/ui"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	// enable once you have at least one migration
	_ "pocketvue/migrations"
)

// Initialize environment variables
func init() {
	godotenv.Load()
	if err := config.Init(); err != nil {
		log.Printf("Warning: Failed to initialize config: %v", err)
	}
}

func main() {
	app := pocketbase.New()

	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Dashboard
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})

	// Register hooks
	hooks.RegisterUserCreatedHook(app)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(ui.DistDirFS, true)).
			BindFunc(func(e *core.RequestEvent) error {
				if e.Request.URL.Path != "/" {
					e.Response.Header().Set("Cache-Control", "max-age=1209600, stale-while-revalidate=86400")
				}
				return e.Next()
			}).
			Bind(apis.Gzip())
		se.Router.GET("/api/workspaces", routes.GetAllWorkspaces) // this is a test endpoint
		se.Router.GET("/api/products", routes.GetProducts)
		se.Router.POST("/api/checkout", routes.CreateCheckoutSession)
		se.Router.POST("/api/customer-portal", routes.CreateCustomerPortalSession)
		se.Router.POST("/api/polar-webhook", routes.HandlePolarWebhook)
		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

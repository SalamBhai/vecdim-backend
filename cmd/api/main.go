package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/salambhai/vecdim-backend/internal/alerts"
	"github.com/salambhai/vecdim-backend/internal/auth"
	"github.com/salambhai/vecdim-backend/internal/dashboard"
	"github.com/salambhai/vecdim-backend/internal/dhis2"
	"github.com/salambhai/vecdim-backend/internal/docs"
	"github.com/salambhai/vecdim-backend/internal/pipeline"
	"github.com/salambhai/vecdim-backend/internal/platform/config"
	"github.com/salambhai/vecdim-backend/internal/platform/database"
	"github.com/salambhai/vecdim-backend/internal/platform/response"
	"github.com/salambhai/vecdim-backend/internal/records"
	"github.com/salambhai/vecdim-backend/internal/spatial"
	"github.com/salambhai/vecdim-backend/internal/sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Initializing VECDIM Modular Monolith API...")

	// Load environment variables / .env configuration
	appCfg := config.Load()

	// Database connection (PostGIS / PostgreSQL)
	dbCfg := database.Config{
		DatabaseURL: appCfg.DatabaseURL,
		Host:        appCfg.DBHost,
		Port:        appCfg.DBPort,
		User:        appCfg.DBUser,
		Password:    appCfg.DBPassword,
		DBName:      appCfg.DBName,
		SSLMode:     appCfg.DBSSLMode,
	}
	client, rawDB, err := database.Connect(ctx, dbCfg)
	if err != nil {
		log.Printf("PostGIS connection warning: %v (running in mock/standalone mode)", err)
	} else {
		defer client.Close()
		log.Println("Connected to PostGIS 16-3.4 database successfully!")

		// Auto-seed if database is empty (e.g. fresh cloud deployment)
		if alertCount, countErr := client.Alert.Query().Count(ctx); countErr == nil && alertCount == 0 {
			log.Println("Database is empty. Automatically seeding demo users, alerts, and risk points...")
			_ = database.SeedAll(ctx, client)
		}
	}

	// Module Services
	syncService := sync.NewService(client)
	recordsService := records.NewService(client)
	alertsService := alerts.NewService(client)
	spatialService := spatial.NewService(client, rawDB)
	pipelineRunner := pipeline.NewRunner(client)

	// Handlers
	syncHandler := sync.NewHandler(syncService)
	recordsHandler := records.NewHandler(recordsService)
	alertsHandler := alerts.NewHandler(alertsService)
	spatialHandler := spatial.NewHandler(spatialService)
	dhis2Handler := dhis2.NewHandler(client)
	pipelineHandler := pipeline.NewHandler(pipelineRunner)
	docsHandler := docs.NewHandler()
	dashboardHandler := dashboard.NewHandler()

	// Background Cron Worker (Configurable interval from .env)
	cronWorker := pipeline.NewCronScheduler(pipelineRunner, appCfg.CronInterval())
	cronWorker.Start(ctx)
	defer cronWorker.Stop()

	// Chi Router configuration
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration for React Native mobile app & Web Dashboard
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Id", "X-User-Role", "X-User-State"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Auth Middleware
	r.Use(auth.Middleware)

	// Decision-Support GIS Web Dashboard & Swagger Documentation
	r.Get("/", dashboardHandler.ServeDashboard)
	r.Get("/dashboard", dashboardHandler.ServeDashboard)
	r.Get("/docs", docsHandler.ServeUI)
	r.Get("/swagger", docsHandler.ServeUI)
	r.Get("/api/openapi.json", docsHandler.ServeSpec)

	// Register Routes under /api/v1 and root for maximum client compatibility
	registerRoutes := func(router chi.Router) {
		// Health Check (must)
		router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusOK, response.Envelope{"status": "ok"})
		})

		// Sync / Ingestion (must)
		router.Post("/records/batch", syncHandler.HandleBatchSync)

		// Records query (must)
		router.Get("/records", recordsHandler.HandleGetRecords)

		// Alerts (must & bonus acknowledge)
		router.Get("/alerts", alertsHandler.HandleGetAlerts)
		router.Post("/alerts/{id}/acknowledge", alertsHandler.HandleAcknowledgeAlert)

		// Spatial / Risk Points (bonus)
		router.Get("/risk-points", spatialHandler.HandleGetRiskPoints)

		// DHIS2 Export (bonus)
		router.Get("/export/dhis2", dhis2Handler.HandleExportDHIS2)

		// R Pipeline Execution Trigger (on-demand orchestration)
		router.Post("/pipeline/run", pipelineHandler.HandleRunPipeline)
	}

	// Register both /api/v1 and root paths
	r.Route("/api/v1", registerRoutes)
	registerRoutes(r)

	port := appCfg.Port
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown listener
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("VECDIM API Server listening on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server stopped unexpectedly: %v", err)
		}
	}()

	<-shutdown
	log.Println("Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}

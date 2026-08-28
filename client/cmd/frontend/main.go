package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/client"
	"github.com/Damione1/thread-art-generator/client/internal/handlers"
	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get working directory")
	}

	staticDir := filepath.Join(workDir, "client/public")
	if _, err := os.Stat("/app/client/public"); err == nil {
		staticDir = "/app/client/public"
	}

	log.Info().Str("staticDir", staticDir).Msg("Using static files directory")

	dbDSN := config.GetPostgresDSN()

	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}

	sessionManager, err := auth.NewSCSSessionManager(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create SCS session manager")
	}

	httpClient := &http.Client{
		Transport: client.NewSessionTransport(sessionManager),
	}

	artGeneratorClient := pbconnect.NewArtGeneratorServiceClient(
		httpClient,
		config.ApiURL,
	)

	generatorService := services.NewGeneratorService(artGeneratorClient, sessionManager)
	passwordAuth := handlers.NewPasswordAuthHandler(&coreauth.PGIdentities{DB: db}, sessionManager)
	pageHandler := handlers.NewPageHandler(generatorService)
	artHandler := handlers.NewArtHandler(generatorService)
	compositionHandler := handlers.NewCompositionHandler(generatorService)

	r := chi.NewRouter()

	r.Use(sessionManager.GetSessionManager().LoadAndSave)
	r.Use(client.IncomingCookieMiddleware)
	r.Use(middleware.SessionAuthMiddleware(sessionManager))

	r.Group(func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", passwordAuth.Login)
			r.Post("/signup", passwordAuth.Signup)
			r.Post("/logout", passwordAuth.Logout)
			r.Get("/logout", passwordAuth.Logout)
			r.Get("/status", passwordAuth.Status)
		})

		r.Get("/", pageHandler.HomePage)
		r.Get("/login", pageHandler.LoginPage)
		r.Get("/signup", pageHandler.SignupPage)

		mountRPCProxy(r, config.ApiURL)
	})

	r.Group(func(r chi.Router) {
		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/", pageHandler.DashboardPage)
			r.Route("/arts", func(r chi.Router) {
				r.Get("/new", artHandler.NewArtPage)
				r.Post("/new", artHandler.CreateArt)
				r.Get("/{artId}", artHandler.ViewArtPage)

				r.Route("/{artId}/composition", func(r chi.Router) {
					r.Get("/new", compositionHandler.NewCompositionForm)
					r.Post("/new", compositionHandler.CreateComposition)
					r.Get("/{compositionId}", compositionHandler.ViewComposition)
					r.Delete("/{compositionId}", compositionHandler.DeleteComposition)
				})
			})
		})

		r.Route("/api", func(r chi.Router) {
			r.Get("/user", func(w http.ResponseWriter, r *http.Request) {
				user, ok := middleware.UserFromContext(r.Context())

				w.Header().Set("Content-Type", "application/json")
				if !ok || user == nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"error":"Unauthorized"}`))
					return
				}

				fmt.Fprintf(w, `{"id":"%s","name":"%s","email":"%s"}`,
					user.ID, user.Name, user.Email)
			})
		})
	})

	fileServer := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	srv := &http.Server{
		Addr:    ":" + config.FrontendPort,
		Handler: r,
	}

	go func() {
		log.Info().Str("port", config.FrontendPort).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server shutdown failed")
	}

	log.Info().Msg("Server gracefully stopped")
}

func mountRPCProxy(r chi.Router, apiURLRaw string) {
	if apiURLRaw == "" {
		apiURLRaw = "http://api:9090"
	}
	apiURL, err := url.Parse(apiURLRaw)
	if err != nil {
		log.Error().Err(err).Str("apiURL", apiURLRaw).Msg("invalid ApiURL, /rpc proxy not mounted")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(apiURL)
	proxy.FlushInterval = -1
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.URL.Host = apiURL.Host
		req.URL.Scheme = apiURL.Scheme
		req.Host = apiURL.Host
	}

	r.Mount("/rpc", http.StripPrefix("/rpc", proxy))
	log.Info().Str("apiURL", apiURL.String()).Msg("/rpc reverse proxy mounted")
}

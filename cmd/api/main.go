package main

import (
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/rs/zerolog/log"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/cache"
	database "github.com/Damione1/thread-art-generator/core/db"
	"github.com/Damione1/thread-art-generator/core/interceptors"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
	"github.com/Damione1/thread-art-generator/core/service"
	"github.com/Damione1/thread-art-generator/core/util"
)

func main() {
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to load config")
	}

	_, err = database.ConnectDb(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to connect to database")
	}

	go cache.CleanExpiredCacheEntries()
	runConnectServer(config)
}

func runConnectServer(config util.Config) {
	log.Print("🍩 Starting Connect server...")
	server, err := service.NewServer(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}
	defer server.Close()
	log.Print("🍩 Server created")

	authService, err := createAuthService(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize auth service")
	}

	var hmacAuth auth.ServiceAuth
	hmacAuth, err = auth.NewHMACServiceAuth(config.ServiceHMACSecret)
	if err != nil {
		log.Warn().Err(err).Msg("HMAC service auth disabled (secret too short); cookie/Firebase still work")
		hmacAuth = nil
	}

	sm := scs.New()
	sm.Store = postgresstore.New(config.DB)
	sm.Cookie.Name = "session_id"
	sm.Cookie.HttpOnly = true
	sm.Cookie.SameSite = http.SameSiteLaxMode
	sm.Lifetime = 24 * time.Hour
	sessions, err := auth.NewSCSSessions(sm)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize SCS sessions")
	}

	interceptorChain := connect.WithInterceptors(
		interceptors.ConnectLogger(),
		interceptors.IdentityInterceptor(sessions, hmacAuth),
		interceptors.AuthMiddleware(authService),
	)

	path, handler := pbconnect.NewArtGeneratorServiceHandler(server, interceptorChain)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle(path, handler)

	reflector := grpcreflect.NewStaticReflector(pbconnect.ArtGeneratorServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	checker := grpchealth.NewStaticChecker(pbconnect.ArtGeneratorServiceName)
	mux.Handle(grpchealth.NewHandler(checker))

	serverPort := config.HTTPServerPort
	if serverPort == "" {
		serverPort = config.GRPCServerPort
		log.Warn().Msg("HTTP_SERVER_PORT not set, using GRPC_SERVER_PORT instead")
	}
	addr := fmt.Sprintf("0.0.0.0:%s", serverPort)
	log.Print("🍩 Starting to listen on " + addr)

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	srv := &http.Server{
		Addr:              addr,
		Handler:           interceptors.HttpLogger(mux),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		Protocols:         protocols,
	}
	log.Fatal().Err(srv.ListenAndServe()).Msg("failed to listen")
}

func createAuthService(config util.Config) (auth.AuthService, error) {
	firebaseConfig := auth.FirebaseConfiguration{
		ProjectID:    config.Firebase.ProjectID,
		EmulatorHost: config.Firebase.EmulatorHost,
	}

	return auth.NewFirebaseAuthService(firebaseConfig)
}

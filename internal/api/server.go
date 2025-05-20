package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mmd-moradi/goup/configs"
	"github.com/mmd-moradi/goup/internal/auth"
	customMiddleware "github.com/mmd-moradi/goup/internal/middleware"
	repositories "github.com/mmd-moradi/goup/internal/repository"
	"github.com/mmd-moradi/goup/internal/repository/postgres"
	"github.com/mmd-moradi/goup/internal/service"
	"github.com/mmd-moradi/goup/internal/storage"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var WhiteListOrigins = map[string]bool{
	"http://localhost:3000": true,
	"https://mmdmidev.site": true,
}

type Server struct {
	*http.Server
	router     chi.Router
	logger     zerolog.Logger
	tokenSvc   *auth.TokenService
	userSvc    *service.UserService
	photoSvc   *service.PhotoService
	storageSvc storage.StorageService
	userRepo   repositories.UserRepository
	photoRepo  repositories.PhotoRepository
}

func NewServer(
	cfg *configs.Config,
	logger zerolog.Logger,
) *Server {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Use(cors.Handler(cors.Options{
		AllowOriginFunc:  AllowOriginFunc,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Use(customMiddleware.RequestLogger(logger))

	server := &Server{
		Server: &http.Server{
			Addr:         cfg.Server.Addr,
			Handler:      router,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
		logger: logger,
		router: router,
	}

	if err := server.InitDependencies(cfg); err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize dependencies")
	}

	server.routes()
	return server
}

func (s *Server) InitDependencies(cfg *configs.Config) error {
	db, err := postgres.NewDBPool(&cfg.Database, s.logger)
	if err != nil {
		return err
	}

	s.userRepo = postgres.NewUserRepository(db)
	s.photoRepo = postgres.NewPhotoRepository(db)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	s.tokenSvc = auth.NewTokenService(redisClient, &cfg.Auth)

	s.storageSvc, err = storage.NewS3StorageService(&cfg.AWS, s.logger)
	if err != nil {
		return err
	}

	s.userSvc = service.NewUserService(s.userRepo, s.tokenSvc, s.logger)
	s.photoSvc = service.NewPhotoService(s.photoRepo, s.userRepo, s.storageSvc, s.logger)

	return nil
}

func (s *Server) routes() {

	userHandler := NewUserHandler(s.userSvc)
	photoHandler := NewPhotoHandler(s.photoSvc)

	authMiddleware := customMiddleware.Authenticate(s.tokenSvc)

	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	s.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Route("/auth", func(r chi.Router) {
				userHandler.RegisterRoutes(r, authMiddleware)
			})
			r.Route("/photos", func(r chi.Router) {
				photoHandler.RegisterRoutes(r, authMiddleware)
			})
		})
	})
}

func AllowOriginFunc(r *http.Request, origin string) bool {
	if _, ok := WhiteListOrigins[origin]; ok {
		return true
	}
	return false
}

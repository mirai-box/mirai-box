package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"

	"github.com/mirai-box/mirai-box/internal/config"
	"github.com/mirai-box/mirai-box/internal/database"
	"github.com/mirai-box/mirai-box/internal/handler"
	"github.com/mirai-box/mirai-box/internal/repository"
	"github.com/mirai-box/mirai-box/internal/service"
)

var corsConfig = cors.New(cors.Options{
	AllowedOrigins:   []string{"http://localhost:3000"}, // Allow frontend origin
	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	ExposedHeaders:   []string{"Link"},
	AllowCredentials: true,
	MaxAge:           300,
})

// Application contains the components of your application.
type Application struct {
	Router http.Handler
	DB     *sqlx.DB
}

// New initializes the application with all its dependencies.
func New(conf *config.Config) (*Application, error) {
	conn, err := database.NewConnection(conf)
	if err != nil {
		return nil, err
	}

	userRepo := repository.NewSQLUserRepository(conn)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService, conf.SessionKey)

	pictureRepo := repository.NewPictureRepository(conn)
	storageRepo := repository.NewStorageRepository(conf.StorageRoot)

	pictureManagementService := service.NewPictureManagementService(pictureRepo, storageRepo)
	pictureRetrievalService := service.NewPictureRetrievalService(pictureRepo, storageRepo)

	pictureRetrievalHandler := handler.NewPictureRetrievalHandler(pictureRetrievalService)
	pictureManagementHandler := handler.NewPictureManagementHandler(pictureManagementService)

	galleryRepo := repository.NewSQLGalleryRepository(conn)
	galleryService := service.NewGalleryService(galleryRepo)
	galleryHandler := handler.NewGalleryHandler(galleryService)

	webPageRepo := repository.NewWebPageRepository(conn)
	webPageService := service.NewWebPageService(webPageRepo)
	webPageHandler := handler.NewWebPageHandler(webPageService)

	mux := chi.NewRouter()

	// A good base middleware stack
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(corsConfig.Handler)

	// Public Routes
	mux.Get("/art/{artID}", pictureRetrievalHandler.SharedPictureHandler)
	mux.Get("/galleries/main", galleryHandler.GetMainGallery)
	mux.Get("/galleries/{galleryID}/images", galleryHandler.GetImagesByGalleryIDHandler)

	// Admin Routes
	mux.Route("/stash", func(r chi.Router) {
		r.Use(userHandler.AuthMiddleware) // Middleware for protecting admin routes

		r.Get("/pictures", pictureManagementHandler.ListPicturesHandler)
		r.Get("/pictures/{pictureID}", pictureRetrievalHandler.LatestFileDownloadHandler)
		r.Get("/pictures/{pictureID}/revisions", pictureManagementHandler.ListRevisionHandler)
		r.Get("/pictures/{pictureID}/revisions/{revisionID}", pictureRetrievalHandler.FileRevisionDownloadHandler)
		r.Post("/pictures/{pictureID}/revisions", pictureManagementHandler.AddRevisionHandler)
		r.Post("/pictures/upload", pictureManagementHandler.UploadHandler)

		r.Get("/galleries", galleryHandler.ListGalleries)
		r.Post("/galleries", galleryHandler.CreateGallery)
		r.Post("/galleries/{galleryID}/images", galleryHandler.AddImageToGallery)
		r.Post("/galleries/{galleryID}/publish", galleryHandler.PublishGallery)

		r.Get("/webpages", webPageHandler.ListWebPagesHandler)
		r.Get("/webpages/{id}", webPageHandler.GetWebPageHandler)
		r.Post("/webpages", webPageHandler.CreateWebPageHandler)
		r.Put("/webpages/{id}", webPageHandler.UpdateWebPageHandler)
		r.Delete("/webpages/{id}", webPageHandler.DeleteWebPageHandler)
	})

	mux.Post("/login", userHandler.LoginHandler)
	mux.Get("/auth/check", userHandler.AuthCheckHandler)

	return &Application{
		Router: mux,
		DB:     conn,
	}, nil
}

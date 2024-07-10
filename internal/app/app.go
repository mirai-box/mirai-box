package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"github.com/rs/cors"
	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/config"
	"github.com/mirai-box/mirai-box/internal/handlers"
	authmiddleware "github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/repos"
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

func SetupRoutes(db *gorm.DB, conf *config.Config) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsConfig.Handler)

	// Initialize repositories
	storageRepo := repos.NewStorageRepository(conf.StorageRoot)
	userRepo := repos.NewUserRepository(db)
	stashRepo := repos.NewStashRepository(db)
	artProjectRepo := repos.NewArtProjectRepository(db)
	revisionRepo := repos.NewRevisionRepository(db)
	collectionRepo := repos.NewCollectionRepository(db)
	collectionArtProjectRepo := repos.NewCollectionArtProjectRepository(db)
	saleRepo := repos.NewSaleRepository(db)
	storageUsageRepo := repos.NewStorageUsageRepository(db)
	webPageRepo := repos.NewWebPageRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	stashService := service.NewStashService(stashRepo)
	artProjectService := service.NewArtProjectService(artProjectRepo, stashRepo)
	revisionService := service.NewRevisionService(revisionRepo)
	collectionService := service.NewCollectionService(collectionRepo)
	collectionArtProjectService := service.NewCollectionArtProjectService(collectionArtProjectRepo)
	saleService := service.NewSaleService(saleRepo)
	storageUsageService := service.NewStorageUsageService(storageUsageRepo)
	webPageService := service.NewWebPageService(webPageRepo)
	artProjectManagementService := service.NewArtProjectManagementService(artProjectRepo, storageRepo, stashRepo)
	cookieStore := sessions.NewCookieStore([]byte(conf.SessionKey))
	m := authmiddleware.NewMiddleware(cookieStore, userService)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, stashService, cookieStore)
	stashHandler := handlers.NewStashHandler(stashService)
	artProjectHandler := handlers.NewArtProjectHandler(artProjectService)
	artProjectManagementHandler := *handlers.NewArtProjectManagementHandler(artProjectManagementService, artProjectService)
	revisionHandler := handlers.NewRevisionHandler(revisionService)
	collectionHandler := handlers.NewCollectionHandler(collectionService)
	collectionArtProjectHandler := handlers.NewCollectionArtProjectHandler(collectionArtProjectService)
	saleHandler := handlers.NewSaleHandler(saleService)
	storageUsageHandler := handlers.NewStorageUsageHandler(storageUsageService)
	webPageHandler := handlers.NewWebPageHandler(webPageService)

	r.Post("/login", userHandler.Login)
	r.Get("/login/check", userHandler.LoginCheck)

	r.Route("/self", func(r chi.Router) {
		r.Use(m.AuthMiddleware)
		r.Use(m.RequireRole("self", "any"))

		r.Get("/stash", stashHandler.MyStash)
		r.Get("/sales", saleHandler.MySales)

		r.Post("/webpages", webPageHandler.CreateWebPage)
		r.Get("/webpages", webPageHandler.MyWebPages)
		r.Get("/webpages/{id}", webPageHandler.MyWebPageByID)
		r.Put("/webpages/{id}", webPageHandler.UpdateWebPage)

		r.Get("/artprojects", artProjectManagementHandler.MyArtProjects)
		r.Get("/artprojects/{id}", artProjectManagementHandler.MyArtProjectByID)
		r.Post("/artprojects/{id}/revision", artProjectManagementHandler.AddRevision)
		r.Post("/artprojects", artProjectManagementHandler.CreateArtProject)

		// r.Get("/collections", collectionHandler.MyCollections)
		// r.Get("/storage", storageUsageHandler.MyStorageUsage)
	})

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Post("/users", userHandler.CreateUser)

		r.Group(func(r chi.Router) {
			r.Use(m.RequireRole("any"))
			r.Post("/webpages", webPageHandler.CreateWebPage)
		})

		r.Group(func(r chi.Router) {
			r.Use(m.RequireRole("admin"))

			// User routes
			r.Get("/users/{id}", userHandler.GetUser)
			r.Put("/users/{id}", userHandler.UpdateUser)
			r.Delete("/users/{id}", userHandler.DeleteUser)

			// Stash routes
			r.Get("/stashes/{id}", stashHandler.FindByID)
			r.Get("/users/{userId}/stash", stashHandler.FindByUserID)

			// ArtProject routes
			r.Post("/artprojects", artProjectHandler.CreateArtProject)
			r.Get("/artprojects/{id}", artProjectHandler.GetArtProject)
			r.Put("/artprojects/{id}", artProjectHandler.UpdateArtProject)
			r.Delete("/artprojects/{id}", artProjectHandler.DeleteArtProject)
			r.Get("/stashes/{stashId}/artprojects", artProjectHandler.ListStashArtProjects)

			// Revision routes
			r.Post("/artprojects/{artProjectId}/revisions", revisionHandler.CreateRevision)
			r.Get("/revisions/{id}", revisionHandler.GetRevision)
			r.Get("/artprojects/{artProjectId}/revisions", revisionHandler.ListArtProjectRevisions)

			// Collection routes
			r.Post("/collections", collectionHandler.CreateCollection)
			r.Get("/collections/{id}", collectionHandler.GetCollection)
			r.Put("/collections/{id}", collectionHandler.UpdateCollection)
			r.Delete("/collections/{id}", collectionHandler.DeleteCollection)
			r.Get("/users/{userId}/collections", collectionHandler.ListUserCollections)

			// CollectionArtProject routes
			r.Post("/collections/{collectionId}/artprojects/{artProjectId}", collectionArtProjectHandler.AddArtProjectToCollection)
			r.Delete("/collections/{collectionId}/artprojects/{artProjectId}", collectionArtProjectHandler.RemoveArtProjectFromCollection)
			r.Get("/collections/{collectionId}/artprojects", collectionArtProjectHandler.ListCollectionArtProjects)

			// Sale routes
			r.Post("/sales", saleHandler.CreateSale)
			r.Get("/sales/{id}", saleHandler.GetSale)
			r.Get("/users/{userId}/sales", saleHandler.ListUserSales)
			r.Get("/artprojects/{artProjectId}/sales", saleHandler.ListArtProjectSales)

			// StorageUsage routes
			r.Get("/users/{userId}/storage", storageUsageHandler.GetUserStorageUsage)
			r.Put("/users/{userId}/storage", storageUsageHandler.UpdateUserStorageUsage)

			// WebPage routes
			// r.Post("/webpages", webPageHandler.CreateWebPage)
			r.Get("/webpages/{id}", webPageHandler.GetWebPage)
			r.Delete("/webpages/{id}", webPageHandler.DeleteWebPage)
			r.Get("/users/{userId}/webpages", webPageHandler.ListUserWebPages)
		})
	})

	return r
}

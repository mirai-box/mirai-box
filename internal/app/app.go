package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"github.com/rs/cors"
	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/config"
	"github.com/mirai-box/mirai-box/internal/handler"
	am "github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/repo"
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
	ur := repo.NewUserRepository(db)
	ar := repo.NewArtProjectRepository(db)
	fsr := repo.NewFileStorageRepository(db, conf.StorageRoot)
	wpr := repo.NewWebPageRepository(db)

	// Initialize services
	userService := service.NewUserService(ur)
	artProjectService := service.NewArtProjectService(ur, ar, fsr, conf.SecretKey)
	webPageService := service.NewWebPageService(wpr)

	cookieStore := sessions.NewCookieStore([]byte(conf.SessionKey))
	m := am.NewMiddleware(cookieStore, userService)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService, cookieStore)
	artProjectHandler := handler.NewArtProjectHandler(artProjectService)
	webPageHandler := handler.NewWebPageHandler(webPageService)

	r.Post("/login", userHandler.Login)
	r.Get("/login/check", userHandler.LoginCheck)
	r.Get("/art/{artID}", artProjectHandler.GetArtByID)

	r.Route("/self", func(r chi.Router) {
		r.Use(m.AuthMiddleware)
		r.Use(m.RequireRole("self", "any"))

		r.Get("/stash", userHandler.MyStash)
		r.Post("/webpages", webPageHandler.CreateWebPage)
		r.Get("/webpages", webPageHandler.MyWebPages)
		r.With(am.ValidateUUID("id")).Get("/webpages/{id}", webPageHandler.MyWebPageByID)
		r.With(am.ValidateUUID("id")).Put("/webpages/{id}", webPageHandler.UpdateWebPage)
		r.With(am.ValidateUUID("id")).Delete("/webpages/{id}", webPageHandler.DeleteWebPage)

		r.Get("/artprojects", artProjectHandler.MyArtProjects)
		r.With(am.ValidateUUID("id")).Get("/artprojects/{id}", artProjectHandler.MyArtProjectByID)

		r.With(am.ValidateUUID("id")).
			Get("/artprojects/{id}/revisions", artProjectHandler.ListRevisions)
		r.With(am.ValidateUUID("id")).
			Post("/artprojects/{id}/revisions", artProjectHandler.AddRevision)
		r.Post("/artprojects", artProjectHandler.CreateArtProject)
		r.With(am.ValidateUUID("artID")).With(am.ValidateUUID("revisionID")).
			Get("/artprojects/{artID}/revisions/{revisionID}", artProjectHandler.RevisionDownload)
	})

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		// user registration route
		r.Post("/users", userHandler.CreateUser)

		r.Group(func(r chi.Router) {
			// r.Use(m.RequireRole("admin"))
		})
	})

	return r
}

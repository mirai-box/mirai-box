package main

import (
	"log/slog"
	"net/http/cgi"

	"github.com/mirai-box/mirai-box/internal/app"
	"github.com/mirai-box/mirai-box/internal/config"
	"github.com/mirai-box/mirai-box/internal/logger"
)

func main() {
	conf, err := config.GetApplicationConfig()
	if err != nil {
		slog.Error("Can't get application config", "error", err)
		return
	}
	logger.Setup(conf)

	// Check if running in a CGI environment
	if conf.IsCGI {
		runCGI(conf)
	} else {
		panic("Not a CGI environment")
	}
}

func runCGI(conf *config.Config) {
	app, err := app.New(conf)
	if err != nil {
		slog.Error("Can't initialize application", "error", err, "CGI", conf.IsCGI)
		panic(err)
	}
	defer app.DB.Close()

	// Handle requests via CGI
	if err := cgi.Serve(app.Router); err != nil {
		slog.Error("Failed to start CGI handler", "error", err)
	}
}

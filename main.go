package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ahowes/passkey-go/config"
	"github.com/ahowes/passkey-go/db"
	"github.com/ahowes/passkey-go/handlers"
	"github.com/ahowes/passkey-go/middleware"
)

func main() {
	cfg := config.Load()

	database := db.Connect(cfg.DatabaseDSN, cfg.DatabaseUser, cfg.DatabasePassword)
	defer database.Close()

	if err := db.CreateTables(context.Background(), database); err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.RPDisplayName,
		RPID:          cfg.RPID,
		RPOrigins:     cfg.RPOrigins,
	})
	if err != nil {
		log.Fatalf("failed to initialize webauthn: %v", err)
	}

	authHandler := handlers.NewAuthHandler(wa, database)
	pageHandler := handlers.NewPageHandler(database)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("passkey_session", store))

	r.GET("/", pageHandler.Index)
	r.GET("/register", pageHandler.Register)
	r.GET("/login", pageHandler.Login)

	api := r.Group("/api")
	{
		api.POST("/register/begin", authHandler.BeginRegistration)
		api.POST("/register/finish", authHandler.FinishRegistration)
		api.POST("/login/begin", authHandler.BeginLogin)
		api.POST("/login/finish", authHandler.FinishLogin)
	}

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		protected.GET("/dashboard", pageHandler.Dashboard)
	}

	r.POST("/logout", authHandler.Logout)

	log.Printf("Starting server on %s", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

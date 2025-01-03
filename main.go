package main

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/url"

	"github.com/benjasper/releases.one/internal/config"
	"github.com/benjasper/releases.one/internal/repository"
	"github.com/benjasper/releases.one/internal/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

//go:embed frontend/dist/*
var distFS embed.FS

func DistFS() *fs.FS {
	efs, err := fs.Sub(distFS, "frontend/dist")
	if err != nil {
		panic(fmt.Sprintf("unable to serve frontend: %v", err))
	}
	return &efs
}

func main() {
	// We ignore the error because we don't want to crash the program if the .env file is missing
	_ = godotenv.Load()
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.IsProduction {
		log.Println("Starting in production mode")
	} else {
		log.Println("Starting in development mode")
	}

	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		log.Fatalf("BASE_URL must be set and must be a valid URL: %s", err)
	}

	githubCallbackURL := fmt.Sprintf("%s/github", baseURL.String())

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GithubClientID,
		ClientSecret: cfg.GithubClientSecret,
		RedirectURL:  githubCallbackURL,
		Scopes:       []string{"user:email", "read:user"}, // Add scopes as needed
		Endpoint:     github.Endpoint,
	}

	db, err := sql.Open("mysql", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %s", err)
	}

	repository := repository.New(db)

	server := server.NewServer(cfg, repository, oauthConfig, baseURL, DistFS())
	server.Start()
}

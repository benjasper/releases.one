package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/benjasper/releases.one/pkg/repository"
	"github.com/benjasper/releases.one/pkg/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func main() {
	// We ignore the error because we don't want to crash the program if the .env file is missing
	_ = godotenv.Load()

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURL := os.Getenv("GITHUB_CALLBACK_URL")

	var (
		oauthConfig = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"user:email", "read:user"}, // Add scopes as needed
			Endpoint:     github.Endpoint,
		}
	)

	connectionString := os.Getenv("DATABASE_URL")

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	repository := repository.New(db)

	server := server.NewServer(repository, oauthConfig)
	server.Start()
}

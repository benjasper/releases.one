package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/benjasper/releases.one/internal/repository"
	"github.com/benjasper/releases.one/internal/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func main() {
	// We ignore the error because we don't want to crash the program if the .env file is missing
	_ = godotenv.Load()

	baseURLString := os.Getenv("BASE_URL")
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		log.Fatal("BASE_URL must be set and must be a valid URL")
	}

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	githubCallbackURL := fmt.Sprintf("%s/github", baseURL.String())

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  githubCallbackURL,
		Scopes:       []string{"user:email", "read:user"}, // Add scopes as needed
		Endpoint:     github.Endpoint,
	}

	connectionString := os.Getenv("DATABASE_URL")

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	repository := repository.New(db)

	server := server.NewServer(repository, oauthConfig, baseURL)
	server.Start()
}

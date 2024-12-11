package repository

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
)

type GitHubToken oauth2.Token

// Implementing `sql.Scanner` interface to read JSON from the database
func (j *GitHubToken) Scan(value interface{}) error {
	if value == nil {
		*j = GitHubToken{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON: %v", value)
	}
	return json.Unmarshal(bytes, j)
}

// Implementing `driver.Valuer` interface to write JSON to the database
func (j GitHubToken) Value() (driver.Value, error) {
	return json.Marshal(j)
}

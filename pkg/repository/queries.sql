-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: CreateUser :execresult
INSERT INTO users (username, refresh_token) VALUES (?, ?);

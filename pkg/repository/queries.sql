-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: CreateUser :execresult
INSERT INTO users (username, refresh_token) VALUES (?, ?);

-- name: GetRepositoryByName :one
SELECT * FROM repositories WHERE name = ?;

-- name: CreateRepository :exec
INSERT INTO repositories (name, url, private, created_at, updated_at, last_synced_at) VALUES (?, ?, ?, ?, ?, ?);

-- name: InsertRepositoryStar :exec
INSERT INTO repository_stars (repository_id, user_id, created_at, updated_at) VALUES (?, ?, ?, ?);

-- name: UpdateRepositoryStar :execresult
UPDATE repository_stars SET updated_at = ? WHERE repository_id = ? AND user_id = ?;

-- name: DeleteRepositoryStarsUpdatedBefore :execresult
DELETE FROM repository_stars WHERE updated_at < ? AND user_id = ?;

-- name: GetReleases :many
SELECT * FROM releases WHERE repository_id = ? ORDER BY released_at DESC;

-- name: InsertRelease :exec
INSERT INTO releases (repository_id, tag_name, description, released_at, created_at, updated_at, is_prerelease) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: DeleteLastXReleases :execresult
DELETE FROM releases WHERE id IN (SELECT id FROM releases AS r WHERE r.repository_id = ? ORDER BY r.released_at DESC LIMIT ?);

-- name: GetReleasesForUser :many
SELECT `releases`.* FROM `releases` INNER JOIN `repository_stars` ON `releases`.`repository_id` = `repository_stars`.`repository_id` WHERE `repository_stars`.`user_id` = ?;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: CreateUser :execresult
INSERT INTO users (username, github_token, last_synced_at) VALUES (?, ?, ?);

-- name: UpdateUserToken :exec
UPDATE users SET github_token = ? WHERE id = ?;

-- name: UpdateUserSyncedAt :exec
UPDATE users SET last_synced_at = ? WHERE id = ?;

-- name: GetUsersInNeedOfAnUpdate :many
SELECT * FROM users WHERE last_synced_at < ?;

-- name: GetRepositoryByName :one
SELECT * FROM repositories WHERE name = ?;

-- name: CreateRepository :exec
INSERT INTO repositories (name, url, image_url, private, created_at, updated_at, last_synced_at) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: InsertRepositoryStar :exec
INSERT INTO repository_stars (repository_id, user_id, created_at, updated_at) VALUES (?, ?, ?, ?);

-- name: UpdateRepositoryStar :execresult
UPDATE repository_stars SET updated_at = ? WHERE repository_id = ? AND user_id = ?;

-- name: DeleteRepositoryStarsUpdatedBefore :execresult
DELETE FROM repository_stars WHERE updated_at < ? AND user_id = ?;

-- name: GetReleases :many
SELECT * FROM releases WHERE repository_id = ? ORDER BY released_at DESC;

-- name: InsertRelease :exec
INSERT INTO releases (repository_id, name, tag_name, url, description, released_at, created_at, updated_at, is_prerelease) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteLastXReleases :execresult
DELETE FROM releases WHERE id IN (SELECT id FROM releases AS r WHERE r.repository_id = ? ORDER BY r.released_at DESC LIMIT ?);

-- name: GetReleasesForUser :many
SELECT `releases`.*, `repositories`.`name` AS repository_name FROM `releases` LEFT JOIN `repositories` ON `releases`.`repository_id` = `repositories`.`id` INNER JOIN `repository_stars` ON `releases`.`repository_id` = `repository_stars`.`repository_id` WHERE `repository_stars`.`user_id` = ? ORDER BY releases.released_at DESC;

-- name: GetUserByGitHubID :one
SELECT
  *
FROM
  users
WHERE
  github_id = ?;

-- name: GetUserByID :one
SELECT
  *
FROM
  users
WHERE
  id = ?;

-- name: CreateUser :execresult
INSERT INTO
  users (github_id, username, github_token, last_synced_at)
VALUES
  (?, ?, ?, ?);

-- name: UpdateUserToken :exec
UPDATE users
SET
  github_token = ?
WHERE
  id = ?;

-- name: UpdateUserPublicID :exec
UPDATE users
SET
  public_id = ?
WHERE
  id = ?;

-- name: UpdateUserSyncedAt :exec
UPDATE users
SET
  last_synced_at = ?
WHERE
  id = ?;

-- name: GetUserByPublicID :one
SELECT
  *
FROM
  users
WHERE
  public_id = ?;

-- name: GetUsersInNeedOfAnUpdate :many
SELECT
  *
FROM
  users
WHERE
  last_synced_at < ?;

-- name: GetRepositoryByGithubID :one
SELECT
  *
FROM
  repositories
WHERE
  github_id = ?;

-- name: CreateRepository :exec
INSERT INTO
  repositories (
    github_id,
    name,
    url,
    image_url,
    image_size,
    private,
    created_at,
    updated_at,
    last_synced_at,
    hash
  )
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateRepository :execresult
UPDATE repositories
SET
  url = ?,
  image_url = ?,
  image_size = ?,
  private = ?,
  created_at = ?,
  updated_at = ?,
  last_synced_at = ?,
  hash = ?
WHERE
  id = ?;

-- name: InsertRepositoryStar :exec
INSERT INTO
  repository_stars (repository_id, user_id, created_at, updated_at)
VALUES
  (?, ?, ?, ?);

-- name: UpdateRepositoryStar :execresult
UPDATE repository_stars
SET
  updated_at = ?
WHERE
  repository_id = ?
  AND user_id = ?;

-- name: DeleteRepositoryStarsUpdatedBefore :execresult
DELETE FROM repository_stars
WHERE
  updated_at < ?
  AND user_id = ?;

-- name: GetReleases :many
SELECT
  *
FROM
  releases
WHERE
  repository_id = ?
ORDER BY
  released_at DESC;

-- name: InsertRelease :exec
INSERT INTO
  releases (
    github_id,
    repository_id,
    name,
    author,
    tag_name,
    url,
    description,
    description_short,
    hash,
    released_at,
    created_at,
    updated_at,
    is_prerelease
  )
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateRelease :execresult
UPDATE releases
SET
  github_id = ?,
  name = ?,
  url = ?,
  description = ?,
  description_short = ?,
  author = ?,
  is_prerelease = ?,
  released_at = ?,
  updated_at = ?,
  hash = ?
WHERE
  id = ?;

-- name: DeleteReleasesOlderThan :execresult
DELETE FROM releases
WHERE
  released_at < ?
  AND repository_id = ?
ORDER BY
  released_at DESC;

-- name: GetReleasesForUser :many
SELECT
  `releases`.`id`,
  `releases`.`repository_id`,
  `releases`.`name`,
  `releases`.`url`,
  `releases`.`tag_name`,
  `releases`.`description`,
  `releases`.`author`,
  `releases`.`is_prerelease`,
  `releases`.`released_at`,
  `releases`.`created_at`,
  `releases`.`updated_at`,
  `repositories`.`name` AS repository_name,
  `repositories`.`image_url` AS image_url,
  `repositories`.`image_size` AS image_size
FROM
  `releases`
  LEFT JOIN `repositories` ON `releases`.`repository_id` = `repositories`.`id`
  INNER JOIN `repository_stars` ON `releases`.`repository_id` = `repository_stars`.`repository_id`
WHERE
  `repository_stars`.`user_id` = ?
ORDER BY
  releases.released_at DESC
LIMIT
  100;

-- name: GetReleasesForUserShortDescription :many
SELECT
  `releases`.`id`,
  `releases`.`repository_id`,
  `releases`.`name`,
  `releases`.`url`,
  `releases`.`tag_name`,
  `releases`.`description_short`,
  `releases`.`author`,
  `releases`.`is_prerelease`,
  `releases`.`released_at`,
  `releases`.`created_at`,
  `releases`.`updated_at`,
  `repositories`.`name` AS repository_name,
  `repositories`.`image_url` AS image_url,
  `repositories`.`image_size` AS image_size
FROM
  `releases`
  LEFT JOIN `repositories` ON `releases`.`repository_id` = `repositories`.`id`
  INNER JOIN `repository_stars` ON `releases`.`repository_id` = `repository_stars`.`repository_id`
WHERE
  `repository_stars`.`user_id` = ?
ORDER BY
  releases.released_at DESC
LIMIT
  100;

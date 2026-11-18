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
  users (
    github_id,
    username,
    github_token,
    last_synced_at,
    is_public,
	is_onboarded,
    public_id
  )
VALUES
  (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateUserToken :exec
UPDATE users
SET
  github_token = ?
WHERE
  id = ?;

-- name: UpdateUserIsPublic :exec
UPDATE users
SET
  is_public = ?
WHERE
  id = ?;

-- name: UpdateUserOnboarded :exec
UPDATE users
SET
  is_onboarded = ?
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
  last_synced_at < ?
ORDER BY
  last_synced_at DESC
LIMIT
  ?;

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

-- name: FindRepositoriesByUser :many
SELECT
  *
FROM
  repositories
LEFT JOIN
  repository_stars ON repositories.id = repository_stars.repository_id
WHERE
  user_id = ?;

-- name: InsertRepositoryStar :exec
INSERT INTO
  repository_stars (repository_id, user_id, type, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?);

-- name: UpdateRepositoryStar :execresult
UPDATE repository_stars
SET
  updated_at = ?
WHERE
  repository_id = ?
  AND user_id = ?
  AND type = ?;

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
SELECT DISTINCT
  `releases`.`id`,
  `releases`.`github_id`,
  `releases`.`repository_id`,
  `releases`.`name`,
  `releases`.`url`,
  `releases`.`tag_name`,
  `releases`.`description`,
  `releases`.`description_short`,
  `releases`.`author`,
  `releases`.`is_prerelease`,
  `releases`.`released_at`,
  `releases`.`created_at`,
  `releases`.`updated_at`,
  `repositories`.`name` AS repository_name,
  `repositories`.`image_url` AS image_url,
  `repositories`.`image_size` AS image_size,
  `repositories`.`github_id` AS repository_github_id,
  `repositories`.`url` AS repository_url
FROM
  `releases`
  LEFT JOIN `repositories` ON `releases`.`repository_id` = `repositories`.`id`
  INNER JOIN `repository_stars` ON `releases`.`repository_id` = `repository_stars`.`repository_id`
  INNER JOIN `users` ON `repository_stars`.`user_id` = `users`.`id`
WHERE
  `repository_stars`.`user_id` = ?
  AND `users`.`is_public` = true
  AND (sqlc.narg('is_prerelease') IS NULL OR `is_prerelease` = sqlc.narg('is_prerelease'))
  AND (sqlc.narg('star_type') IS NULL OR `repository_stars`.`type` = sqlc.narg('star_type'))
  AND (sqlc.narg('list_id') IS NULL OR EXISTS (
    SELECT 1 FROM `star_list_repositories`
    INNER JOIN `star_lists` ON `star_list_repositories`.`star_list_id` = `star_lists`.`id`
    WHERE `star_list_repositories`.`repository_id` = `releases`.`repository_id`
      AND `star_lists`.`id` = sqlc.narg('list_id')
      AND `star_lists`.`user_id` = `repository_stars`.`user_id`
  ))
ORDER BY
  releases.released_at DESC
LIMIT
  100;

-- name: GetReleasesForUserShortDescription :many
SELECT
  `releases`.`id`,
  `releases`.`github_id`,
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
  `repositories`.`image_size` AS image_size,
  `repositories`.`url` AS repository_url,
  MIN(`repository_stars`.`type`) AS repository_star_type
FROM
  `releases`
  LEFT JOIN `repositories` ON `releases`.`repository_id` = `repositories`.`id`
  INNER JOIN `repository_stars` ON `releases`.`repository_id` = `repository_stars`.`repository_id`
WHERE
  `repository_stars`.`user_id` = ?
  AND (sqlc.narg('is_prerelease') IS NULL OR `is_prerelease` = sqlc.narg('is_prerelease'))
  AND (sqlc.narg('star_type') IS NULL OR `repository_stars`.`type` = sqlc.narg('star_type'))
  AND (sqlc.narg('list_id') IS NULL OR EXISTS (
    SELECT 1 FROM `star_list_repositories`
    INNER JOIN `star_lists` ON `star_list_repositories`.`star_list_id` = `star_lists`.`id`
    WHERE `star_list_repositories`.`repository_id` = `releases`.`repository_id`
      AND `star_lists`.`id` = sqlc.narg('list_id')
      AND `star_lists`.`user_id` = `repository_stars`.`user_id`
  ))
GROUP BY
  `releases`.`id`,
  `releases`.`github_id`,
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
  `repositories`.`name`,
  `repositories`.`image_url`,
  `repositories`.`image_size`,
  `repositories`.`url`
ORDER BY
  releases.released_at DESC
LIMIT
  100;

-- name: UpsertStarList :exec
INSERT INTO
  star_lists (github_id, user_id, name, slug, is_private, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  slug = VALUES(slug),
  is_private = VALUES(is_private),
  updated_at = VALUES(updated_at);

-- name: GetStarListByGithubID :one
SELECT
  *
FROM
  star_lists
WHERE
  github_id = ?;

-- name: GetStarListsForUser :many
SELECT
  *
FROM
  star_lists
WHERE
  user_id = ?
ORDER BY
  name ASC;

-- name: UpsertStarListRepository :exec
INSERT INTO
  star_list_repositories (star_list_id, repository_id, updated_at)
SELECT
  sqlc.arg('star_list_id'), `repositories`.`id`, sqlc.arg('updated_at')
FROM
  repositories
WHERE
  `repositories`.`github_id` = sqlc.arg('repository_github_id')
ON DUPLICATE KEY UPDATE
  updated_at = VALUES(updated_at);

-- name: DeleteStarListRepositoriesUpdatedBefore :execresult
DELETE FROM star_list_repositories
WHERE
  `star_list_repositories`.`updated_at` < sqlc.arg('updated_at')
  AND `star_list_repositories`.`star_list_id` IN (
    SELECT
      `star_lists`.`id`
    FROM
      star_lists
    WHERE
      `star_lists`.`user_id` = sqlc.arg('user_id')
  );

-- name: DeleteStarListsUpdatedBefore :execresult
DELETE FROM star_lists
WHERE
  updated_at < ?
  AND user_id = ?;

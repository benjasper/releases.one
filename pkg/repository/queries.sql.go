// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package repository

import (
	"context"
	"database/sql"
	"time"
)

const createRepository = `-- name: CreateRepository :exec
INSERT INTO repositories (name, url, image_url, private, created_at, updated_at, last_synced_at) VALUES (?, ?, ?, ?, ?, ?, ?)
`

type CreateRepositoryParams struct {
	Name         string
	Url          string
	ImageUrl     string
	Private      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastSyncedAt time.Time
}

func (q *Queries) CreateRepository(ctx context.Context, arg CreateRepositoryParams) error {
	_, err := q.db.ExecContext(ctx, createRepository,
		arg.Name,
		arg.Url,
		arg.ImageUrl,
		arg.Private,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.LastSyncedAt,
	)
	return err
}

const createUser = `-- name: CreateUser :execresult
INSERT INTO users (username, github_token, last_synced_at) VALUES (?, ?, ?)
`

type CreateUserParams struct {
	Username     string
	GithubToken  GitHubToken
	LastSyncedAt time.Time
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, createUser, arg.Username, arg.GithubToken, arg.LastSyncedAt)
}

const deleteReleasesOlderThan = `-- name: DeleteReleasesOlderThan :execresult
DELETE FROM releases WHERE released_at < ? AND repository_id = ? ORDER BY released_at DESC
`

type DeleteReleasesOlderThanParams struct {
	ReleasedAt   time.Time
	RepositoryID int32
}

func (q *Queries) DeleteReleasesOlderThan(ctx context.Context, arg DeleteReleasesOlderThanParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, deleteReleasesOlderThan, arg.ReleasedAt, arg.RepositoryID)
}

const deleteRepositoryStarsUpdatedBefore = `-- name: DeleteRepositoryStarsUpdatedBefore :execresult
DELETE FROM repository_stars WHERE updated_at < ? AND user_id = ?
`

type DeleteRepositoryStarsUpdatedBeforeParams struct {
	UpdatedAt time.Time
	UserID    int32
}

func (q *Queries) DeleteRepositoryStarsUpdatedBefore(ctx context.Context, arg DeleteRepositoryStarsUpdatedBeforeParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, deleteRepositoryStarsUpdatedBefore, arg.UpdatedAt, arg.UserID)
}

const getReleases = `-- name: GetReleases :many
SELECT id, repository_id, name, url, tag_name, description, author, is_prerelease, released_at, created_at, updated_at FROM releases WHERE repository_id = ? ORDER BY released_at DESC
`

func (q *Queries) GetReleases(ctx context.Context, repositoryID int32) ([]Release, error) {
	rows, err := q.db.QueryContext(ctx, getReleases, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Release
	for rows.Next() {
		var i Release
		if err := rows.Scan(
			&i.ID,
			&i.RepositoryID,
			&i.Name,
			&i.Url,
			&i.TagName,
			&i.Description,
			&i.Author,
			&i.IsPrerelease,
			&i.ReleasedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getReleasesForUser = `-- name: GetReleasesForUser :many
SELECT releases.id, releases.repository_id, releases.name, releases.url, releases.tag_name, releases.description, releases.author, releases.is_prerelease, releases.released_at, releases.created_at, releases.updated_at, ` + "`" + `repositories` + "`" + `.` + "`" + `name` + "`" + ` AS repository_name, ` + "`" + `repositories` + "`" + `.` + "`" + `image_url` + "`" + ` AS image_url FROM ` + "`" + `releases` + "`" + ` LEFT JOIN ` + "`" + `repositories` + "`" + ` ON ` + "`" + `releases` + "`" + `.` + "`" + `repository_id` + "`" + ` = ` + "`" + `repositories` + "`" + `.` + "`" + `id` + "`" + ` INNER JOIN ` + "`" + `repository_stars` + "`" + ` ON ` + "`" + `releases` + "`" + `.` + "`" + `repository_id` + "`" + ` = ` + "`" + `repository_stars` + "`" + `.` + "`" + `repository_id` + "`" + ` WHERE ` + "`" + `repository_stars` + "`" + `.` + "`" + `user_id` + "`" + ` = ? ORDER BY releases.released_at DESC
`

type GetReleasesForUserRow struct {
	ID             int32
	RepositoryID   int32
	Name           string
	Url            string
	TagName        string
	Description    string
	Author         sql.NullString
	IsPrerelease   bool
	ReleasedAt     time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	RepositoryName sql.NullString
	ImageUrl       sql.NullString
}

func (q *Queries) GetReleasesForUser(ctx context.Context, userID int32) ([]GetReleasesForUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getReleasesForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetReleasesForUserRow
	for rows.Next() {
		var i GetReleasesForUserRow
		if err := rows.Scan(
			&i.ID,
			&i.RepositoryID,
			&i.Name,
			&i.Url,
			&i.TagName,
			&i.Description,
			&i.Author,
			&i.IsPrerelease,
			&i.ReleasedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.RepositoryName,
			&i.ImageUrl,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRepositoryByName = `-- name: GetRepositoryByName :one
SELECT id, name, url, private, created_at, updated_at, last_synced_at, image_url FROM repositories WHERE name = ?
`

func (q *Queries) GetRepositoryByName(ctx context.Context, name string) (Repository, error) {
	row := q.db.QueryRowContext(ctx, getRepositoryByName, name)
	var i Repository
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Url,
		&i.Private,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastSyncedAt,
		&i.ImageUrl,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, username, github_token, last_synced_at FROM users WHERE username = ?
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.GithubToken,
		&i.LastSyncedAt,
	)
	return i, err
}

const getUsersInNeedOfAnUpdate = `-- name: GetUsersInNeedOfAnUpdate :many
SELECT id, username, github_token, last_synced_at FROM users WHERE last_synced_at < ?
`

func (q *Queries) GetUsersInNeedOfAnUpdate(ctx context.Context, lastSyncedAt time.Time) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, getUsersInNeedOfAnUpdate, lastSyncedAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.GithubToken,
			&i.LastSyncedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertRelease = `-- name: InsertRelease :exec
INSERT INTO releases (repository_id, name, author, tag_name, url, description, released_at, created_at, updated_at, is_prerelease) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

type InsertReleaseParams struct {
	RepositoryID int32
	Name         string
	Author       sql.NullString
	TagName      string
	Url          string
	Description  string
	ReleasedAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsPrerelease bool
}

func (q *Queries) InsertRelease(ctx context.Context, arg InsertReleaseParams) error {
	_, err := q.db.ExecContext(ctx, insertRelease,
		arg.RepositoryID,
		arg.Name,
		arg.Author,
		arg.TagName,
		arg.Url,
		arg.Description,
		arg.ReleasedAt,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.IsPrerelease,
	)
	return err
}

const insertRepositoryStar = `-- name: InsertRepositoryStar :exec
INSERT INTO repository_stars (repository_id, user_id, created_at, updated_at) VALUES (?, ?, ?, ?)
`

type InsertRepositoryStarParams struct {
	RepositoryID int32
	UserID       int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (q *Queries) InsertRepositoryStar(ctx context.Context, arg InsertRepositoryStarParams) error {
	_, err := q.db.ExecContext(ctx, insertRepositoryStar,
		arg.RepositoryID,
		arg.UserID,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const updateRepositoryStar = `-- name: UpdateRepositoryStar :execresult
UPDATE repository_stars SET updated_at = ? WHERE repository_id = ? AND user_id = ?
`

type UpdateRepositoryStarParams struct {
	UpdatedAt    time.Time
	RepositoryID int32
	UserID       int32
}

func (q *Queries) UpdateRepositoryStar(ctx context.Context, arg UpdateRepositoryStarParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, updateRepositoryStar, arg.UpdatedAt, arg.RepositoryID, arg.UserID)
}

const updateUserSyncedAt = `-- name: UpdateUserSyncedAt :exec
UPDATE users SET last_synced_at = ? WHERE id = ?
`

type UpdateUserSyncedAtParams struct {
	LastSyncedAt time.Time
	ID           int32
}

func (q *Queries) UpdateUserSyncedAt(ctx context.Context, arg UpdateUserSyncedAtParams) error {
	_, err := q.db.ExecContext(ctx, updateUserSyncedAt, arg.LastSyncedAt, arg.ID)
	return err
}

const updateUserToken = `-- name: UpdateUserToken :exec
UPDATE users SET github_token = ? WHERE id = ?
`

type UpdateUserTokenParams struct {
	GithubToken GitHubToken
	ID          int32
}

func (q *Queries) UpdateUserToken(ctx context.Context, arg UpdateUserTokenParams) error {
	_, err := q.db.ExecContext(ctx, updateUserToken, arg.GithubToken, arg.ID)
	return err
}

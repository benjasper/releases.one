# Releases.one 🧵🛜

An app to provide you with a feed of releases from your starred and subscribed GitHub repositories.

## Features

- Get a feed (atom or rss) of releases from your starred and subscribed GitHub repositories
- View the timeline of releases in the frontend on releases.one
- Filter out prereleases and whether to use your starred or subscribed repositories

## How does it work (technically)

You login with your GitHub which gives the app access to your starred and subscribed repositories (and nothing else).

The app then uses the GitHub GraphQL API (with your token) to get the releases for each repository. Each request fetches 50 repositories at a time.
One request consumes one point of [GitHubs rate limit tokens](https://docs.github.com/en/graphql/overview/rate-limits-and-node-limits-for-the-graphql-api#primary-rate-limit) (not much, you have 5000 per hour).
It keeps the releases in a database and resyncs your list every 2 hours.
That means you will have the latest releases with a maximum delay of 2 hours. Bonus: Users are not synced at the same time.
So the more users have the same repos in their lists the more frequent the update interval gets.

## Tech stack

- Go
    - Connect (GRPC)
    - sqlc
- SolidJS with [Solid UI](https://www.solid-ui.com/)
- MySQL

## How to run it in development

This project uses [Task](https://taskfile.dev/) to run the commands.

1. `docker-compose up -d` to start the database
2. `migrations-apply` to apply the db schema
2. `task dev` to start the api and frontend

## How to host it yourself

1. You can use the [Dockerfile](./Dockerfile) to build the app yourself
2. An example of all the necessary environment variables is in [.env.example](./.env.example)
3. You need to [create a GitHub OAuth app](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app) and set the `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` environment variables.
    - The callback URL should be `YOUR_DOMAIN/api/callback`
    - You also want to select the permissions for "Starring" and "Watching" under "Permissions & events"

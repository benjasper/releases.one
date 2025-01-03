FROM node:20-slim AS node-builder

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
COPY ./frontend /frontend
WORKDIR /frontend

FROM node-builder AS prod-deps
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod --frozen-lockfile

FROM node-builder AS build
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm run build

FROM golang:latest AS go-builder
WORKDIR /app
COPY . .
COPY --from=prod-deps /frontend/dist /app/frontend/dist
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/releases ./main.go

FROM alpine:latest AS runtime
COPY --from=go-builder /bin/releases /bin/releases
CMD ["/bin/releases"]




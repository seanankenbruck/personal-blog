# Personal Blog

[![CodeQL Advanced](https://github.com/seanankenbruck/personal-blog/actions/workflows/codeql.yml/badge.svg)](https://github.com/seanankenbruck/personal-blog/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/seanankenbruck/personal-blog/graph/badge.svg?token=30DNBFU5QA)](https://codecov.io/gh/seanankenbruck/personal-blog)

A modern, test-driven personal blog built with Go. Posts are rendered from static Markdown files. The app is containerized and can be deployed to any kubernetes cluster via the deployment scripts in `./deploy` or published to Azure App Service via GitHub Actions.

## Features

- HTTP server using Gin
- Posts sourced from Markdown in `content/posts`
- Server-side Markdown rendering to HTML templates
- Simple file-based content workflow (commit a `.md` to publish)
- Docker image published to Docker Hub
- GitHub Actions pipelines for app and infra

## Prerequisites

- Go 1.24.5 or later
- Docker (for building/running the container)

## Getting Started (Local)

1. Clone the repository
2. Install Go dependencies:
   ```bash
   go mod download && go mod tidy
   ```
3. Run the application:
   ```bash
   go run cmd/main.go
   ```
4. Visit `http://localhost:8080`

### Adding a Post

Add a Markdown file under `content/posts` with front matter in the filename `YYYY-MM-DD-my-post.md`. The server discovers posts at startup via the file repository.

## Docker

The Dockerfile is multi-stage. It supports dynamic architecture using `ARG TARGETARCH` with a default of `amd64`.

Build locally:
```bash
docker build -t smankenb/personal-blog:local .
```

On Apple Silicon or Raspberry Pi (arm64):
```bash
docker build --platform linux/arm64 \
  --build-arg TARGETARCH=arm64 \
  -t smankenb/personal-blog:arm64 .
```

## Testing

```bash
go test ./...
```

## Deployment Overview

- Infrastructure (resource group, App Service, etc.) is managed with Pulumi in `infra/`
- App deployment builds and pushes a Docker image, then updates the Azure App Service
- GitHub Environments (`dev`, `prod`) hold per-environment secrets such as `AZURE_APP_SERVICE_NAME` and `AZURE_APP_SERVICE_RESOURCE_GROUP`

See `infra/README.md` for infrastructure instructions and GitHub Actions for the CI/CD flow.

## Project Structure

```
.
├── cmd/                # Application entry point
├── content/           # Markdown posts
├── internal/          # Application code (config, handlers, services)
├── static/            # Static assets
├── templates/         # HTML templates
├── infra/             # Pulumi IaC for Azure
└── .github/workflows/ # CI/CD pipelines
```

## License

MIT
# JobJaldi Backend

This directory contains the Go backend for JobJaldi.

## Structure

- `cmd/api`: The HTTP API server.
- `cmd/scraper`: The CLI tool for scraping jobs.
- `internal/jobs`: Job models and database logic.
- `internal/scrapers`: Scraper implementations.

## Running Locally

### API

```bash
cd backend
go run ./cmd/api
```
The API will start on port 8080.

### Scraper

```bash
cd backend
go run ./cmd/scraper
```

## Deployment

This project is designed to be deployed on **Render** (API) and **GitHub Actions** (Scraper).

# Internship Tracker

Internship Tracker is a REST API written in Go for keeping track of internship and job applications.

The project allows applications, companies, notes, contacts and follow-ups to be stored in PostgreSQL and managed through HTTP endpoints.

The application is containerized with Docker and Docker Compose and includes database migrations, tests and CI with GitHub Actions.

## Tech Stack
- Go
- Chi Router
- PostgreSQL
- pgx
- Docker
- Docker Compose
- Goose migrations
- GitHub Actions

## Requirements

To run the project with Docker you need:

- Docker
- Docker Compose

To run Go commands locally you also need Go installed.

## Getting Started

Clone the repository:

```bash
git clone https://github.com/Teddinator/Internship-tracker
cd Internship-tracker
```

Create the required environment configuration if it does not already exist.

The application expects a PostgreSQL connection through:

`DATABASE_URL`

Build and start the application and PostgreSQL database:

`make dev`

The api is available at:
`http://localhost:8080`

The PostgreSQL container is exposed locally on port:

5433

To stop the containers:

`docker compose down`

## Makefile

The project contains a Makefile for common development commands.

Start the development environment:
`make dev`

Run database migrations:
`make migrate-up`

Roll back the latest migration:
`make migrate-down`

Create a database backup:
`make backup-db`

Create seed data for application:
`make seed`

Run unit tests in container:
`make test`

## Health Check

Check that the API is running:

`curl http://localhost:8080/health`

Example response:

```JSON
{
  "status": "ok"
}
```

## API Endpoints

### Health

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Check API health |

### Applications

| Method | Endpoint | Description |
|---|---|---|
| GET | `/applications/` | Get all applications |
| POST | `/applications/` | Create an application |
| GET | `/applications/{id}` | Get an application by ID |
| PUT | `/applications/{id}` | Update an application |
| DELETE | `/applications/{id}` | Delete an application |
| GET | `/applications/export.csv` | Export applications as CSV |
| GET | `/applications/{id}/notes` | Get notes for an application |
| POST | `/applications/{id}/notes` | Create a note for an application |
| POST | `/applications/{id}/followups` | Create a follow-up for an application |

### Companies

| Method | Endpoint | Description |
|---|---|---|
| GET | `/companies/` | Get all companies |
| POST | `/companies/` | Create a company |
| GET | `/companies/{id}` | Get a company by ID |
| PUT | `/companies/{id}` | Update a company |
| DELETE | `/companies/{id}` | Delete a company |

### Notes

| Method | Endpoint | Description |
|---|---|---|
| GET | `/applications/{id}/notes` | Get notes for an application |
| POST | `applications/{id}/notes` | Create a note for an application |

### Follow-ups

| Method | Endpoint | Description |
|---|---|---|
| GET | `/followups/due` | Get due follow-ups |
| PUT | `/followups/{id}/complete` | Mark a follow-up as complete |
| DELETE | `/followups/{id}` | Delete a follow-up |

## Database

The project uses PostgreSQL.

Database schema changes are handled with Goose migrations stored in the `migrations/` directory.

Run migrations with:

`make migrate-up`

Roll back migrations with:

`make migrate-down`

The database is started automatically through Docker Compose.

## Tests

Run all tests:

`make test`

Run tests with verbose output:

`go test -v ./...`

Run Go's static analysis tool:

`go vet ./...`

The project currently includes tests for:

- application status validation
- the /health handler
- company handler validation

## Continuous Integration

GitHub Actions automatically runs checks when code is pushed or a pull request is opened.

The CI workflow runs:

```bash
go vet ./...
go test ./...
```

This verifies that the project passes the same basic checks in a clean GitHub environment and not only on the developer's local machine.

## Database Backup

A compressed PostgreSQL backup can be created with:

`make backup-db` 

Backups are stored in the local `backups/` directory and are not committed to Git.

## Project Structure
.
├── cmd/
│   └── api/
├── internal/
│   ├── applications/
│   ├── companies/
│   ├── contacts/
│   ├── followups/
│   └── notes/
├── migrations/
├── scripts/
├── .github/
│   └── workflows/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md

## Version

Current milestone:

v0.1.0

`v0.1.0` represents the first working MVP of Internship Tracker.
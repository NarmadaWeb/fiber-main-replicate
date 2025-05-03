# Fiber Replicate Example

This project demonstrates a simple application with a main server and a replica server built using the Fiber web framework in Go, with a PostgreSQL database. The main server handles write operations and reads, while the replica server primarily handles read operations from the same database.

## Project Structure

```
.
├── cmd/
│   ├── main_server/
│   │   └── main.go         # Main server entry point
│   └── replica_server/
│       └── main.go       # Replica server entry point
├── internal/
│   ├── config/
│   │   └── config.go     # Application configuration loading
│   ├── handlers/
│   │   ├── common_handlers.go # Handlers used by both servers (e.g., health check)
│   │   ├── items_handlers.go  # Handlers for item operations on the main server
│   │   └── replica_handlers.go # Handlers for item operations on the replica server
│   ├── models/
│   │   └── items.go      # Data models (e.g., Item struct)
│   ├── routers/
│   │   ├── main_routers.go   # Router setup for the main server
│   │   └── replica_routers.go # Router setup for the replica server
│   └── stores/
│       ├── error.go      # Custom error types
│       ├── gorm_stores.go  # GORM implementation of data store
│       └── interface.go    # Data store interface
├── config.yml            # Application configuration file
├── Dockerfile            # Docker build instructions
├── docker-compose.yml    # Docker Compose configuration
├── go.mod                # Go module file
└── go.sum                # Go module checksums
```

## Prerequisites

*   Go (version 1.21 or higher recommended)
*   Docker
*   Docker Compose

## Setup and Running

The recommended way to run this project is using Docker Compose, which will set up the PostgreSQL database, the main server, and the replica server.

1.  **Clone the repository:**
    ```bash
    git clone <repository_url> # Replace <repository_url> with the actual URL
    cd fiber-replicate-example
    ```

2.  **Build and run with Docker Compose:**
    Navigate to the project root directory where `docker-compose.yml` is located and run:
    ```bash
    sudo docker-compose up --build
    ```
    *   `sudo` might be required depending on your Docker installation and user permissions. If you encounter permission errors, ensure your user is in the `docker` group or run with `sudo`.
    *   `--build` will build the Docker image for the servers before starting the containers.

    This command will:
    *   Start a PostgreSQL database container (`fiber_replica_db`).
    *   Build the Go application image.
    *   Start the main server container (`fiber_main_server`) and the replica server container (`fiber_replica_server`).
    *   The servers will be configured to connect to the database service within the Docker network using environment variables defined in `docker-compose.yml`.

3.  **Accessing the services:**
    *   Main Server: `http://localhost:8000`
    *   Replica Server: `http://localhost:8001`

4.  **Stopping the services:**
    To stop the running containers, press `Ctrl+C` in the terminal where `docker-compose up` is running.
    To stop and remove the containers, network, and volumes (except named volumes like `postgres_data`), run:
    ```bash
    sudo docker-compose down
    ```
    To remove named volumes as well (which will delete your database data), use:
    ```bash
    sudo docker-compose down --volumes
    ```

## Configuration

The application uses `config.yml` for configuration. However, when running with Docker Compose, environment variables prefixed with `APP_` (e.g., `APP_DATABASE_MAIN_DSN`, `APP_REPLICASERVER_MAINSERVERADDRESS`) will override the values in `config.yml`.

Key configuration options:

*   `database.main.dsn`: PostgreSQL connection string.
*   `mainServer.port`: Port for the main server.
*   `replicaServer.port`: Port for the replica server.
*   `replicaServer.mainServerAddress`: Address of the main server for the replica to fetch data (used by the `/fetch-from-main` endpoint).

## API Endpoints

**Main Server (`http://localhost:8000`)**

*   `GET /api/v1/health`: Health check.
*   `POST /api/v1/items`: Create a new item.
*   `GET /api/v1/items`: Get all items.
*   `GET /api/v1/items/:id`: Get a specific item by ID.
*   `PUT /api/v1/items/:id`: Update a specific item by ID.
*   `DELETE /api/v1/items/:id`: Delete a specific item by ID.

**Replica Server (`http://localhost:8001`)**

*   `GET /api/v1/health`: Health check.
*   `GET /api/v1/local-data`: Get static local data from the replica.
*   `GET /api/v1/fetch-from-main`: Fetch data from the main server (demonstrates replica fetching).
*   `GET /api/v1/items`: Get all items from the replica's database.
*   `GET /api/v1/items/:id`: Get a specific item by ID from the replica's database.
*   *Note: Write operations (POST, PUT, DELETE) are not supported on the replica server.*

## Database

The project uses PostgreSQL. The main server is responsible for running database migrations on startup. The `docker-compose.yml` sets up a PostgreSQL container and a named volume (`postgres_data`) to persist the database data.

The database DSN is configured via `config.yml` or overridden by the `APP_DATABASE_MAIN_DSN` environment variable.

## Running Locally (Without Docker Compose)

If you prefer to run the servers directly on your host machine, you will need:

1.  A running PostgreSQL server accessible at the DSN specified in `config.yml`.
2.  Run database migrations (this is handled by the main server on startup).
3.  Build and run the main server:
    ```bash
    go build -o main_server ./cmd/main_server/main.go
    ./main_server
    ```
4.  Build and run the replica server:
    ```bash
    go build -o replica_server ./cmd/replica_server/main.go
    ./replica_server
    ```
    Ensure the `replicaServer.mainServerAddress` in `config.yml` is correct if the replica needs to fetch from the main server.

---

This README provides a comprehensive overview and instructions for setting up and running the project.

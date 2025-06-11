# News App - Go CRUD HTMx+MongoDB Application

![News app](assets/screenshots/screenshot-1.png)

## Setup and Run

### Using Docker Compose

1. Clone the repository
2. Run with Docker Compose:

```bash
docker-compose up -d
```

Or using Makefile: 

```bash
make up
```

3. Open your browser and navigate to `http://localhost:8080`
4. Stop the application, remove containers and delete all associated volumes: 

```bash
docker-compose down -v
```

Or using Makefile:

```bash
make remove
```

### Configuration (Optional)

You can configure the application using **Environment variables**: Use uppercase with underscores for nested keys (e.g., `SERVER_PORT=9090`)

#### Configuration Options

| Config File Key | Environment Variable | Default                  | Description                         |
|----------------|----------------------|--------------------------|-------------------------------------|
| server.host | SERVER_HOST          | localhost (or 0.0.0.0 in containers) | Server hostname                     |
| server.port | SERVER_PORT          | 8080                     | Server port number                  |
| mongo.uri | MONGO_URI            | mongodb://localhost:27017 | MongoDB connection URI              |
| mongo.db | MONGO_DB             | news_app                 | MongoDB database name               |
| mongo.timeout | MONGO_TIMEOUT        | 10                       | MongoDB database timeout in seconds |
| app.posts_per_page | APP_POSTS_PER_PAGE   | 12              | Number of posts per page            |
| app.static_directory | APP_STATIC_DIRECTORY | static                   | Directory for static assets          |

## Testing

### Running Tests

Run tests using the following Makefile commands:

```bash
# Run unit tests
make test-unit

# Run integration tests
make test-integration

# Run all tests (unit and integration)
make test-all

# Generate and open coverage report
make coverage

# Clean up coverage report files
make clean
```

## Technologies Used

- **Go 1.24**
- **MongoDB**
- **Chi Router**
- **Viper**
- **HTMx**
- **Tailwind CSS**
- **Docker & Docker Compose**
- **Testify**
- **Dockertest** (for integration testing)

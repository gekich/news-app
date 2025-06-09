# News App - Go CRUD HTMx+MongoDB Application

## Setup and Run

### Using Docker Compose

1. Clone the repository
2. Run with Docker Compose:

```bash
docker-compose up --build
```

Or Makefile: 

```makefile
up
```

3. Open your browser and navigate to `http://localhost:8080`
4. Stop the application, remove containers and delete all associated volumes: 

```bash
docker-compose down -v
```

Or Makefile:

```makefile
remove
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

## Technologies Used

- **Go 1.24**
- **MongoDB**
- **Chi Router**
- **Viper**
- **HTMx**
- **Tailwind CSS**
- **Docker & Docker Compose**

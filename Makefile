# Start the application
up:
	docker-compose up -d

# Stop and remove all running containers
down:
	docker-compose down

# Stop the application, remove containers and delete all associated volumes
remove:
	docker-compose down -v

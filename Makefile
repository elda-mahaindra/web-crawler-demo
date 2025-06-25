# Development environment commands
dev-up:
	docker compose -f docker-compose.dev.yml up -d

dev-down:
	docker compose -f docker-compose.dev.yml down

dev-down-v:
	docker compose -f docker-compose.dev.yml down -v

dev-rm:
	docker rmi web-crawler

# Helper commands
help:
	@echo "Available commands:"
	@echo "  dev-up           - Start development environment in detached mode"
	@echo "  dev-down         - Stop development environment"
	@echo "  dev-down-v - Stop development environment and remove volumes"

.PHONY: dev-up dev-down dev-down-v help

.PHONY: dev prod build down logs restart tidy

dev:
	docker compose up --build -d

prod:
	docker compose -f docker-compose.yml up --build -d

build:
	docker compose -f docker-compose.yml build

down:
	docker compose down

restart:
	docker compose restart

logs:
	docker compose logs -f

tidy:
	go mod tidy

lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run ./...

up:
	docker-compose up -d

db:
	docker-compose up -d auth_pg

down:
	docker-compose down

rebuild:
	docker-compose up -d --remove-orphans --force-recreate --build
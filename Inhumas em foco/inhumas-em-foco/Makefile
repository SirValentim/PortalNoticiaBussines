templ:
	@echo "templ not required for standard Go templates"

dev:
	go run ./cmd/web

worker:
	go run ./cmd/worker

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/web ./cmd/web
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/worker ./cmd/worker
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/seed-news ./cmd/seed-news
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/migrate-sqlite-postgres ./cmd/migrate-sqlite-postgres
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/inhumas-mcp ./cmd/mcp

build-linux: build

mcp-deps:
	go get github.com/mark3labs/mcp-go@v0.32.0
	go mod tidy

mcp-dev:
	DATABASE_URL=$${DATABASE_URL:-./inhumas.db} DB_DRIVER=$${DB_DRIVER:-sqlite} go run ./cmd/mcp

mcp-build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/inhumas-mcp ./cmd/mcp

mcp-build-windows:
	go build -o bin/inhumas-mcp.exe ./cmd/mcp

mcp-inspect:
	DATABASE_URL=$${DATABASE_URL:-./inhumas.db} DB_DRIVER=$${DB_DRIVER:-sqlite} npx @modelcontextprotocol/inspector go run ./cmd/mcp

migrate:
	psql "$$DATABASE_URL" -f migrations/001_init_postgres.sql

deploy: build
	ssh vps "sudo systemctl stop inhumas-web inhumas-worker"
	rsync -avz --delete bin/ vps:/var/www/inhumas/bin/
	rsync -avz static/ vps:/var/www/inhumas/static/
	ssh vps "sudo systemctl start inhumas-web inhumas-worker"

status:
	ssh vps "sudo systemctl status inhumas-web inhumas-worker"

logs:
	ssh vps "sudo journalctl -u inhumas-web -u inhumas-worker -f"

backup:
	ssh vps "sudo BACKUP_DIR=/var/backups/inhumas /var/www/inhumas/scripts/backup.sh"

health:
	curl -s http://localhost:8080/health

smoke:
	BASE_URL=$${BASE_URL:-http://localhost:8080} scripts/smoke-test.sh

readiness:
	BASE_URL=$${BASE_URL:-https://inhumasemfoco.online} scripts/production-readiness.sh

.PHONY: templ dev worker build build-linux mcp-deps mcp-dev mcp-build-linux mcp-build-windows mcp-inspect migrate deploy status logs backup health smoke readiness

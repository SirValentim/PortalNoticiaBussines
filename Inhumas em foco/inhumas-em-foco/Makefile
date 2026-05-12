templ:
	@echo "templ not required for standard Go templates"

dev:
	go run ./cmd/web

worker:
	go run ./cmd/worker

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/web ./cmd/web
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/worker ./cmd/worker

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
	ssh vps "sudo /opt/inhumas/backup.sh"

health:
	curl -s http://localhost:8080/health

.PHONY: templ dev worker build migrate deploy status logs backup health

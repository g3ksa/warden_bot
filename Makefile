COMPOSE := docker-compose -f docker/docker-compose.dev.yml
SITEMAP_POSTGRESQL_DSN := postgres://user:pass@localhost:5435/warden_bot?sslmode=disable

up:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down -v

restart:
	$(COMPOSE) down -v
	$(COMPOSE) up --build -d

ps:
	$(COMPOSE) ps

# EXAMPLE: make name=add_category_and_filter_column goose-new
goose-new:
	goose -dir db/migrations create ${name} sql

goose-up:
	goose -dir db/migrations postgres ${SITEMAP_POSTGRESQL_DSN} up

goose-down:
	goose -dir db/migrations postgres ${SITEMAP_POSTGRESQL_DSN} down

goose-reload:
	goose -dir db/migrations postgres ${SITEMAP_POSTGRESQL_DSN} down
	goose -dir db/migrations postgres ${SITEMAP_POSTGRESQL_DSN} up

service-run:
	go run cmd/warden_bot/main.go
service-build:
	go build -o bin/ diplom/cmd/warden_bot
build-dev:
	docker compose -f docker-compose.yaml -f docker-compose.dev.yaml --env-file .env.dev --env-file backend/.env.dev up -d --build

build-prod:
	docker compose -f docker-compose.yaml -f docker-compose.prod.yaml --env-file .env.prod --env-file backend/.env.prod up -d --build
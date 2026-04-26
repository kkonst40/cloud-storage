# Cloud File Storage

## Конфигурация
Приложение запускается через docker compose.

Для запуска Dev-стенда используется команда:

`docker compose -f docker-compose.yaml -f docker-compose.dev.yaml --env-file .env.dev --env-file backend/.env.dev up -d --build` 

либо, при наличии Makefile: `make build-dev`

Для запуска Prod-стенда используется команда:

`docker compose -f docker-compose.yaml -f docker-compose.prod.yaml --env-file .env.prod --env-file backend/.env.prod up -d --build`

либо, при наличии Makefile: `make build-prod`

Env для Dev-стенда: `.env.dev.example` в корневой папке и в папке `backend/` необходимо переименовать в `.env.dev`.

Env для Prod-стенда: `.env.prod.example` в корневой папке и в папке `backend/` необходимо переименовать в`.env.prod`.

После запуска приложение будет доступно по адресу http://localhost:80
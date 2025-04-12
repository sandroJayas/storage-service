#!/bin/bash
set -e

MODE=${1:-watch}

check_jwt_token_env_var() {
  local var_name=\"$1\"
  local value="${!1}"
  if [[ -z "$value" || "${value: -1}" != "=" ]]; then
    echo "❌ $var_name is not set or does not end with '='"
    exit 1
  fi
}

source "$(dirname "$0")/../.env"
check_jwt_token_env_var JWT_SECRET

echo "🔄 Starting user-service test DB..."
docker run -d --rm --name user-service-db \
  -e POSTGRES_DB=users \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5433:5432 \
  postgres:15-alpine

until psql "postgresql://postgres:postgres@localhost:5433/users" -c '\q' 2>/dev/null; do
  echo "Waiting for Postgres to accept connections..."
  sleep 1
done

echo "📦 Pulling user-service image..."
docker pull sandrojayas/user-service:latest

echo "📄 Loading schema directly from GitHub..."
curl -s https://raw.githubusercontent.com/sandroJayas/user-service/main/migrations/schema.sql | \
  docker run --rm -i --network host \
    postgres:15-alpine \
    psql "postgresql://postgres:postgres@localhost:5433/users" \
    --pset pager=off

echo "🚀 Starting user-service container..."
docker run -d --rm --name user-service --link user-service-db \
  -e DATABASE_URL=postgres://postgres:postgres@user-service-db:5432/users?sslmode=disable \
  -e JWT_SECRET=$JWT_SECRET \
  -e HONEYCOMB_SERVICE_NAME=some-name \
  -e OTEL_EXPORTER_OTLP_ENDPOINT=some-name \
  -e OTEL_EXPORTER_OTLP_HEADERS=some-name \
  -p 8081:8080 \
  sandrojayas/user-service:latest

echo "⏳ Waiting for user-service to become ready..."
for i in {1..15}; do
  if curl -s http://localhost:8081/readyz > /dev/null; then
    echo "✅ user-service is ready!"
    break
  fi
  echo "⏳ waiting... ($i/15)"
  sleep 1
done

# Check final attempt
if ! curl -s http://localhost:8081/readyz > /dev/null; then
  echo "❌ user-service failed to start in time"
  exit 1
fi

if [ "$MODE" = "watch" ]; then
  trap 'echo ""; echo "🧹 Cleaning up..."; docker stop user-service user-service-db; exit 0' INT
  while true; do sleep 1; done
fi

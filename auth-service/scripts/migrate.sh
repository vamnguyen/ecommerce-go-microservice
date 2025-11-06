#!/bin/bash

# Migration script for auth-service database

set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-auth_db}

echo "Running database migrations..."

# Wait for database to be ready
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q'; do
  echo "Waiting for database to be ready..."
  sleep 2
done

echo "Database is ready!"

# Run Go migrations
go run cmd/server/main.go migrate

echo "Migrations completed successfully!"

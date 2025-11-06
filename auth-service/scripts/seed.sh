#!/bin/bash

# Seed database with test data

set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-auth_db}

echo "Seeding database with test data..."

# Create test admin user
PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<EOF
-- Insert admin user (password: Admin@123456)
INSERT INTO users (id, email, password_hash, role, is_verified, is_active, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  'admin@example.com',
  '\$2a\$10\$rKvKzwN5J5J5J5J5J5J5JuqKqKqKqKqKqKqKqKqKqKqKqKqKqKqK',
  'admin',
  true,
  true,
  NOW(),
  NOW()
) ON CONFLICT (email) DO NOTHING;

-- Insert test user (password: Test@123456)
INSERT INTO users (id, email, password_hash, role, is_verified, is_active, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  'test@example.com',
  '\$2a\$10\$rKvKzwN5J5J5J5J5J5J5JuqKqKqKqKqKqKqKqKqKqKqKqKqKqKqK',
  'user',
  true,
  true,
  NOW(),
  NOW()
) ON CONFLICT (email) DO NOTHING;

EOF

echo "Database seeded successfully!"
echo ""
echo "Test users created:"
echo "  Admin: admin@example.com / Admin@123456"
echo "  User:  test@example.com  / Test@123456"

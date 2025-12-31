#!/bin/bash
set -e

echo "Testing database manager setup..."

# Test creating template database
echo "Creating template database..."
psql -U postgres -h localhost -p 5432 -c "DROP DATABASE IF EXISTS storyengine_template;"
psql -U postgres -h localhost -p 5432 -c "CREATE DATABASE storyengine_template;"

# Apply migrations
echo "Applying migrations..."
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/storyengine_template?sslmode=disable" up

echo "✅ Template database created successfully!"

# List tables
echo "Tables in template database:"
psql -U postgres -h localhost -p 5432 -d storyengine_template -c "\dt"

echo "✅ All setup complete!"


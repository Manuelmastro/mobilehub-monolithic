#!/bin/sh
# wait-for-postgres.sh
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done
echo "PostgreSQL is ready!"
exec "$@"  # Execute the original command (the app start)

#!/bin/sh

set -eu

echo "Waiting for PostgreSQL..."

until pg_isready \
    -h "$SQL_HOST" \
    -p "$SQL_PORT" \
    -U "$SQL_USER" \
    -d postgres >/dev/null 2>&1
do
    sleep 1
done

echo "PostgreSQL started"

if [ "${DELETE_DB:-no}" = "yes" ]; then
    echo "Dropping and recreating database..."

    PGPASSWORD="$SQL_PASSWORD" dropdb \
        -h "$SQL_HOST" \
        -p "$SQL_PORT" \
        -U "$SQL_USER" \
        --if-exists \
        --force \
        "$SQL_DATABASE"

    PGPASSWORD="$SQL_PASSWORD" createdb \
        -h "$SQL_HOST" \
        -p "$SQL_PORT" \
        -U "$SQL_USER" \
        "$SQL_DATABASE"
fi

echo "Applying database migrations..."

goose \
    -dir ./database/migrations \
    postgres \
    "host=$SQL_HOST port=$SQL_PORT user=$SQL_USER password=$SQL_PASSWORD dbname=$SQL_DATABASE sslmode=disable" \
    up

exec "$@"
#!/bin/sh

if [ "$DATABASE" = "postgres" ]
then
    echo "Waiting for postgres..."

    while ! nc -z $SQL_HOST $SQL_PORT; do
      sleep 0.1
    done

    echo "PostgreSQL started"

    # For database deletion and recreation. Change the .env.dev file to DELETE_DB=yes to enable this.
    if [ "$DELETE_DB" = "yes" ]; then
        echo "Terminating existing connections to the database..."
        PGPASSWORD=$SQL_PASSWORD psql -h $SQL_HOST -U $SQL_USER -d postgres -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '$SQL_DATABASE' AND pid <> pg_backend_pid();" || echo "Failed to terminate connections"

        echo "Dropping existing database and recreating..."
        PGPASSWORD=$SQL_PASSWORD dropdb -h $SQL_HOST -U $SQL_USER $SQL_DATABASE || echo "Failed to drop database"
        PGPASSWORD=$SQL_PASSWORD createdb -h $SQL_HOST -U $SQL_USER $SQL_DATABASE || echo "Failed to create database"
    else
        echo "Skipping database deletion and recreation."
    fi
fi

# Note: In Go, you will run migrations here using a tool like 'goose' or 'golang-migrate'
# Example: goose -dir ./migrations postgres "user=$SQL_USER password=$SQL_PASSWORD dbname=$SQL_DATABASE host=$SQL_HOST sslmode=disable" up

exec "$@"
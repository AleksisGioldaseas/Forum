#!/bin/sh
set -e

# Initialize DB
if [ ! -f /forum/data/forum.db ]; then
    echo "Initializing database..."
    mkdir -p /forum/data/images
    sqlite3 /forum/data/forum.db < /forum/sql/schema.sql

    # Populate
    if [ "$POPULATE_DB" = "true" ]; then
        echo "Populating test data..."
        /forum/forum-app populate
    fi
fi

# Verify mounts
if [ ! -f /forum/configs.json ]; then
    echo "ERROR: config file must be mounted!" >&2
    exit 1
fi

exec "$@"
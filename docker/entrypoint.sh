#!/bin/sh

set -eu

: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
: "${POSTGRES_DB:?POSTGRES_DB is required}"
: "${TELEGRAM_BOT_TOKEN:?TELEGRAM_BOT_TOKEN is required}"

PGDATA="/var/lib/taskbot-postgres"

mkdir -p "$PGDATA"
mkdir -p /run/postgresql

chown -R postgres:postgres "$PGDATA"
chown -R postgres:postgres /run/postgresql


# ----------------------------------
# 1. Инициализация PostgreSQL
# ----------------------------------

if [ ! -s "$PGDATA/PG_VERSION" ]; then

    echo "Initializing PostgreSQL..."

    su-exec postgres initdb \
        -D "$PGDATA" \
        --username=postgres \
        --auth-local=trust \
        --auth-host=scram-sha-256

    cat >> "$PGDATA/postgresql.conf" <<EOF

listen_addresses = '127.0.0.1'
port = 5432
unix_socket_directories = '/run/postgresql'

EOF
fi


# ----------------------------------
# 2. Запуск PostgreSQL
# ----------------------------------

echo "Starting PostgreSQL..."

su-exec postgres postgres \
    -D "$PGDATA" &

POSTGRES_PID=$!


# ----------------------------------
# 3. Ждём PostgreSQL
# ----------------------------------

until pg_isready \
    -h 127.0.0.1 \
    -p 5432 \
    -U postgres > /dev/null 2>&1
do
    echo "Waiting for PostgreSQL..."
    sleep 1
done

echo "PostgreSQL is ready"


# ----------------------------------
# 4. Создаём пользователя и БД
# ----------------------------------

su-exec postgres psql \
    -v ON_ERROR_STOP=1 \
    --set=db_user="$POSTGRES_USER" \
    --set=db_password="$POSTGRES_PASSWORD" \
    --set=db_name="$POSTGRES_DB" \
    postgres <<'SQL'

SELECT format(
    'CREATE ROLE %I LOGIN PASSWORD %L',
    :'db_user',
    :'db_password'
)
WHERE NOT EXISTS (
    SELECT 1
    FROM pg_roles
    WHERE rolname = :'db_user'
)
\gexec


SELECT format(
    'CREATE DATABASE %I OWNER %I',
    :'db_name',
    :'db_user'
)
WHERE NOT EXISTS (
    SELECT 1
    FROM pg_database
    WHERE datname = :'db_name'
)
\gexec

SQL


# ----------------------------------
# 5. Создаём таблицы
# ----------------------------------

echo "Applying database schema..."

PGPASSWORD="$POSTGRES_PASSWORD" \
psql \
    -h 127.0.0.1 \
    -p 5432 \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -v ON_ERROR_STOP=1 \
    -f /app/schema.sql


# ----------------------------------
# 6. Переменные для Go
# ----------------------------------

export DB_HOST=127.0.0.1
export DB_PORT=5432


# ----------------------------------
# 7. Запуск Go bot
# ----------------------------------

echo "Starting TaskBot..."

/app/taskbot &

APP_PID=$!


# ----------------------------------
# 8. Корректная остановка
# ----------------------------------

shutdown() {

    echo "Stopping TaskBot..."

    kill -TERM "$APP_PID" 2>/dev/null || true
    wait "$APP_PID" 2>/dev/null || true

    echo "Stopping PostgreSQL..."

    su-exec postgres pg_ctl \
        -D "$PGDATA" \
        -m fast \
        -w stop || true
}

trap shutdown TERM INT


wait "$APP_PID"

APP_STATUS=$?

shutdown

exit "$APP_STATUS"
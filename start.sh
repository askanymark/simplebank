#!/usr/bin/env sh

set -e

echo "run db migrations"
source /app/app.env
cat /app/app.env
echo $DB_URI
/app/migrate -path /app/migration -database "$DB_URI" -verbose up

echo "start the app"
exec "$@"
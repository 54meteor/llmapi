#!/bin/bash

set -e

echo "=== Building backend ==="
cd "$(dirname "$0")/one_api"
go build -o one-api .

echo "=== Building frontend ==="
cd "$(dirname "$0")/web/default"
npm run build

echo "=== Copying frontend to web/build ==="
cd "$(dirname "$0")"
mkdir -p web/build/default
cp -r web/default/build/* web/build/default/

echo "=== Build complete ==="
echo "To run the server:"
echo "  SQL_DSN='postgres://user:pass@localhost:5432/one_api' ./one_api --port 3001"

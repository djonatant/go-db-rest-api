#!/bin/bash

# Load environment variables from .env.local if present
if [ -f .env.local ]; then
    export $(grep -v '^#' .env.local | xargs)
fi

# Run the application
# Pass all arguments to the go run command
go run main.go --port=9992 --db-type=mysql --db-host=localhost --db-port=3306 --db-user=user --db-password=user_password --db-name=example_db "$@"

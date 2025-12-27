#!/bin/bash

# Load environment variables
if [ -f .env.local ]; then
    export $(grep -v '^#' .env.local | xargs)
else
    echo ".env.local file not found!"
    exit 1
fi

# Check for AWS CLI
if ! command -v aws &> /dev/null; then
    echo "AWS CLI could not be found. Please install it."
    exit 1
fi

# Configure AWS/R2 credentials for the session
export AWS_ACCESS_KEY_ID=$CF_R2_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY=$CF_R2_SECRET_ACCESS_KEY
export AWS_DEFAULT_REGION="auto"

# Build binaries
mkdir -p bin

echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o bin/app-windows-amd64.exe main.go

echo "Building for Mac (Intel)..."
GOOS=darwin GOARCH=amd64 go build -o bin/app-darwin-amd64 main.go

echo "Building for Mac (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -o bin/app-darwin-arm64 main.go

echo "Building for Linux (Ubuntu amd64)..."
GOOS=linux GOARCH=amd64 go build -o bin/app-linux-amd64 main.go

echo "Build complete."

# Upload to R2
ENDPOINT_URL="https://${CF_ACCOUNT_ID}.r2.cloudflarestorage.com"

echo "Uploading to R2 bucket: $CF_R2_PUBLIC_BUCKET..."
aws s3 cp bin/ s3://$CF_R2_PUBLIC_BUCKET/bin/ --recursive --endpoint-url $ENDPOINT_URL

echo "Deployment complete!"

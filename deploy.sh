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

VERSION=$(date +%Y%m%d%H%M%S)
LDFLAGS="-X main.Version=$VERSION"

echo "Building version $VERSION..."

echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/app-windows-amd64.exe main.go

echo "Building for Mac (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/app-darwin-amd64 main.go

echo "Building for Mac (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o bin/app-darwin-arm64 main.go

echo "Building for Linux (Ubuntu amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/app-linux-amd64 main.go

echo "Build complete."

# Create version.json
cat <<EOF > bin/version.json
{
  "version": "$VERSION",
  "binaries": {
    "windows/amd64": "app-windows-amd64.exe",
    "darwin/amd64": "app-darwin-amd64",
    "darwin/arm64": "app-darwin-arm64",
    "linux/amd64": "app-linux-amd64"
  }
}
EOF

# Upload to R2
if [ -z "$CF_ACCOUNT_ID" ] || [ -z "$CF_R2_PUBLIC_BUCKET" ]; then
    echo "Error: CF_ACCOUNT_ID or CF_R2_PUBLIC_BUCKET is not set!"
    exit 1
fi

ENDPOINT_URL="https://${CF_ACCOUNT_ID}.r2.cloudflarestorage.com"

echo "Uploading to R2 bucket: $CF_R2_PUBLIC_BUCKET..."
# Use 'sync' instead of 'cp --recursive' as it's often more stable on macOS
aws s3 sync bin/ s3://$CF_R2_PUBLIC_BUCKET/bin/ --endpoint-url $ENDPOINT_URL

echo "Deployment complete! Version: $VERSION"

#!/bin/bash

# Default modes
# These could be changed through command line arguments in production
MODE="dev"
IMAGE_NAME="forum-app"
CONTAINER_NAME="forum-container"
PORT=8080

# Paths for bind mounts
DATA_DIR="$(pwd)/data"
CONFIG_FILE="$(pwd)/configs.json"
CERTS_DIR="$(pwd)/certs"

# Ensure directories exist
mkdir -p "$DATA_DIR" "$CERTS_DIR"

# Remove existing containers
echo "Removing old container..."
podman stop "$CONTAINER_NAME" 2>/dev/null || true
podman rm "$CONTAINER_NAME" 2>/dev/null || true

# Build fresh image
make gen-certs

echo "Building Docker image '$IMAGE_NAME'..."
podman build -t "$IMAGE_NAME" .

# Run new container
echo "Running new container '$CONTAINER_NAME' in $MODE mode..."
podman run -d \
    -p "$PORT":8080 \
    -v "$DATA_DIR:/forum/data" \
    -v "$CONFIG_FILE:/forum/configs.json" \
    -v "$CERTS_DIR:/forum/certs" \
    -e "POPULATE_DB=true" \
    --name "$CONTAINER_NAME" \
    "$IMAGE_NAME"
    
    # This is where we could run production variables instead



echo "Forum app running on https://localhost:$PORT"
podman logs -f "$CONTAINER_NAME"
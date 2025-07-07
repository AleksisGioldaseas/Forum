#!/bin/bash

# Config
CONTAINER_NAME="forum-container"
IMAGE_NAME="forum-app"
DATA_DIR="$(pwd)/data"
CERTS_DIR="$(pwd)/certs"

# Stop containers
echo "Removing container '$CONTAINER_NAME'..."
podman stop "$CONTAINER_NAME" 2>/dev/null || true
podman rm "$CONTAINER_NAME" 2>dev/null || true

# Remove all build artifacts
echo "Cleaning up podman resources..."
podman system prune -f

# Remove the image
echo "Removing image '$IMAGE_NAME'..."
podman rmi "$IMAGE_NAME" 2>/dev/null || true

# NUCLEAR OPTION (wipes all untagged images)
# echo "WARNING: Removing all untagged images"
# podman rmi $(podman images -q --filter "dangling=true") 2>/dev/null || true

# Deletes bind mounted data
# rm -rf "${DATA_DIR}" 2>/dev/null || true
# rm -rf "${CERTS_DIR}" 2>/dev/null || true

echo "Cleanup complete"
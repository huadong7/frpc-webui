#!/bin/bash
# Build frpc-manager Docker image for deployment
# Prerequisites: Docker running, Linux binary built at bin/frpc-linux-amd64

set -e

IMAGE_NAME="frpc-manager"
IMAGE_TAG="${1:-latest}"

echo "=== Building frpc-manager Docker image ==="
echo "Image: ${IMAGE_NAME}:${IMAGE_TAG}"

# Build Linux binary if not exists
if [ ! -f "bin/frpc-linux-amd64" ]; then
    echo "Building Linux binary..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -tags frpc -o bin/frpc-linux-amd64 ./cmd/frpc
fi

# Build Docker image
docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" -f dockerfiles/Dockerfile .

echo ""
echo "=== Build complete ==="
echo "Image: ${IMAGE_NAME}:${IMAGE_TAG}"
echo ""
echo "Save for transfer:"
echo "  docker save ${IMAGE_NAME}:${IMAGE_TAG} | gzip > frpc-manager-${IMAGE_TAG}.tar.gz"
echo ""
echo "Load on target machine:"
echo "  docker load < frpc-manager-${IMAGE_TAG}.tar.gz"
echo ""
echo "Run:"
echo "  docker run -d --name frpc-manager -p 7400:7400 -v frpc_data:/data --restart unless-stopped ${IMAGE_NAME}:${IMAGE_TAG}"

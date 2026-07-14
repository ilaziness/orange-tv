#!/bin/bash
set -e

VERSION=${VERSION:-1.0.0}
REGISTRY=${REGISTRY:-docker.io}
IMAGE_NAME=${IMAGE_NAME:-orange-tv}
PUSH=${PUSH:-false}

echo "Deploying orange-tv version ${VERSION}..."

# 构建镜像
docker build -t ${REGISTRY}/${IMAGE_NAME}:${VERSION} \
    --build-arg VERSION=${VERSION} \
    .

# 推送镜像（可选）
if [ "${PUSH}" = "true" ]; then
    echo "Pushing image to registry..."
    docker push ${REGISTRY}/${IMAGE_NAME}:${VERSION}
fi

# 停止旧容器
echo "Stopping old containers..."
docker-compose down

# 启动新容器
echo "Starting new containers..."
docker-compose up -d

echo "Deploy complete"

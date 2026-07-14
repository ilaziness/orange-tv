#!/bin/bash
set -e

VERSION=${VERSION:-1.0.0}
BUILD_DIR=${BUILD_DIR:-build}
GOOS=${GOOS:-linux}
GOARCH=${GOARCH:-amd64}

echo "Building orange-tv version ${VERSION} for ${GOOS}/${GOARCH}..."

# 创建构建目录
mkdir -p ${BUILD_DIR}

# 构建二进制文件
CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
    -ldflags="-s -w -X github.com/ilaziness/orange-tv/cmd.version=${VERSION}" \
    -o ${BUILD_DIR}/orange-tv .

echo "Build complete: ${BUILD_DIR}/orange-tv"

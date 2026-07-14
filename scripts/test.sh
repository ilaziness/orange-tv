#!/bin/bash
set -e

echo "Running tests..."

# 运行单元测试
echo "Running unit tests..."
go test -v ./...

# 运行集成测试
if [ -d "test/integration" ]; then
    echo "Running integration tests..."
    go test -v ./test/integration/...
fi

# 生成覆盖率报告
echo "Generating coverage report..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out | grep total

echo "Tests complete"

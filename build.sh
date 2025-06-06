#!/bin/sh
set -e

echo "开始编译 mowen-mcp..."

# 默认编译 (当前系统和架构)
go build -o bin/mowen-mcp .
echo "默认编译完成: bin/mowen-mcp"

# 交叉编译到各个系统和架构
echo "编译 linux-amd64..."
env GOOS=linux GOARCH=amd64 go build -o bin/mowen-mcp-linux-amd64 .
echo "完成: bin/mowen-mcp-linux-amd64"

echo "编译 linux-386..."
env GOOS=linux GOARCH=386 go build -o bin/mowen-mcp-linux-386 .
echo "完成: bin/mowen-mcp-linux-386"

echo "编译 linux-arm..."
env GOOS=linux GOARCH=arm go build -o bin/mowen-mcp-linux-arm .
echo "完成: bin/mowen-mcp-linux-arm"

echo "编译 linux-arm64..."
env GOOS=linux GOARCH=arm64 go build -o bin/mowen-mcp-linux-arm64 .
echo "完成: bin/mowen-mcp-linux-arm64"

echo "编译 windows-amd64..."
env GOOS=windows GOARCH=amd64 go build -o bin/mowen-mcp-windows-amd64.exe .
echo "完成: bin/mowen-mcp-windows-amd64.exe"

echo "编译 windows-386..."
env GOOS=windows GOARCH=386 go build -o bin/mowen-mcp-windows-386.exe .
echo "完成: bin/mowen-mcp-windows-386.exe"

echo "编译 darwin-amd64..."
env GOOS=darwin GOARCH=amd64 go build -o bin/mowen-mcp-darwin-amd64 .
echo "完成: bin/mowen-mcp-darwin-amd64"

echo "编译 darwin-arm64..."
env GOOS=darwin GOARCH=arm64 go build -o bin/mowen-mcp-darwin-arm64 .
echo "完成: bin/mowen-mcp-darwin-arm64"

echo "所有编译任务已完成!"
echo "二进制文件位于: bin"
#!/bin/bash

# EmojiDB Multi-Platform Build Script
# Created for Robinson Honour

mkdir -p bin/engines

echo "🏗️  Starting EmojiDB Engine Builds..."

# --- MAC (Darwin) ---
echo "🍎 Building for Mac (ARM64)..."
GOOS=darwin GOARCH=arm64 go build -o bin/engines/emojidb-darwin-arm64 ./cmd/emojidb

echo "🍎 Building for Mac (Intel x64)..."
GOOS=darwin GOARCH=amd64 go build -o bin/engines/emojidb-darwin-x64 ./cmd/emojidb

# --- LINUX ---
echo "🐧 Building for Linux (x64)..."
GOOS=linux GOARCH=amd64 go build -o bin/engines/emojidb-linux-x64 ./cmd/emojidb

echo "🐧 Building for Linux (ARM64)..."
GOOS=linux GOARCH=arm64 go build -o bin/engines/emojidb-linux-arm64 ./cmd/emojidb

# --- WINDOWS ---
echo "🪟 Building for Windows (x64)..."
GOOS=windows GOARCH=amd64 go build -o bin/engines/emojidb-win32-x64.exe ./cmd/emojidb

echo "🪟 Building for Windows (ARM64)..."
GOOS=windows GOARCH=arm64 go build -o bin/engines/emojidb-win32-arm64.exe ./cmd/emojidb

echo "✅ All binaries compiled to bin/engines/"
ls -lh bin/engines/

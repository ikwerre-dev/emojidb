#!/bin/bash

# EmojiDB Master Release Script
# Automates: Build -> Tag -> Push -> GitHub Release

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "❌ Error: Please provide a version (e.g., v1.0.4)"
    echo "Usage: ./scripts/release.sh v1.0.0"
    exit 1
fi

# 1. Check for GitHub CLI (gh)
if ! command -v gh &> /dev/null; then
    echo "❌ Error: GitHub CLI (gh) is not installed."
    echo "💡 Install it with: brew install gh"
    echo "   Then run 'gh auth login' to authenticate."
    exit 1
fi

# 2. Build all binaries
echo "🏗️  Phase 1: Compiling multi-platform binaries..."
./scripts/build_all.sh
if [ $? -ne 0 ]; then
    echo "❌ Build failed. Aborting release."
    exit 1
fi

# 3. Create and Push Git Tag
echo "🏷️  Phase 2: Tagging version $VERSION..."
git tag "$VERSION"
git push origin "$VERSION"
if [ $? -ne 0 ]; then
    echo "❌ Failed to push tag. Aborting release."
    exit 1
fi

# 4. Create GitHub Release & Upload Assets
echo "🛰️  Phase 3: Creating GitHub Release and uploading assets..."
gh release create "$VERSION" bin/engines/* --title "Release $VERSION" --notes "EmojiDB Engine binaries for version $VERSION. Includes builds for Mac, Linux, and Windows."

if [ $? -eq 0 ]; then
    echo "✅ SUCCESS! EmojiDB $VERSION is officially released on GitHub."
    echo "🔗 View it at: https://github.com/ikwerre-dev/EmojiDB/releases/tag/$VERSION"
else
    echo "❌ Failed to create GitHub release."
    exit 1
fi

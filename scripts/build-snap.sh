#!/bin/bash

# Build Snap package locally for testing
# Usage: ./scripts/build-snap.sh

set -e

echo "🔨 Building Debian Doctor Snap package..."

# Check if snapcraft is installed
if ! command -v snapcraft &> /dev/null; then
    echo "❌ snapcraft not found. Install with: sudo snap install snapcraft --classic"
    exit 1
fi

# Clean previous builds
echo "🧹 Cleaning previous builds..."
snapcraft clean 2>/dev/null || true

# Get version from git
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
VERSION=${VERSION#v}
echo "📦 Building version: $VERSION"

# Update version in snapcraft.yaml
sed -i.bak "s/version: '1.0.0'/version: '$VERSION'/" snap/snapcraft.yaml

# Build the snap
echo "🏗️  Building snap package..."
snapcraft

# Restore original snapcraft.yaml
mv snap/snapcraft.yaml.bak snap/snapcraft.yaml

# Find the built snap
SNAP_FILE=$(find . -name "debian-doctor_*.snap" | head -1)

if [ -n "$SNAP_FILE" ]; then
    echo "✅ Snap package built successfully: $SNAP_FILE"
    echo ""
    echo "📋 Package info:"
    snap info "$SNAP_FILE" || true
    echo ""
    echo "🔧 To install locally:"
    echo "   sudo snap install --dangerous --classic $SNAP_FILE"
    echo ""
    echo "📤 To test before publishing:"
    echo "   sudo snap install --dangerous --classic $SNAP_FILE"
    echo "   debian-doctor --version"
else
    echo "❌ Snap build failed - no .snap file found"
    exit 1
fi
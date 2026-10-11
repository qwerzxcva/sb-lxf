#!/bin/bash
# sb-lxf Android AAR Build Script
# Builds the Android libbox AAR for sb-AI integration
#
# Usage:
#   ./scripts/build-android-aar.sh [debug]
#
# Environment variables:
#   ANDROID_NDK_HOME - Path to Android NDK (optional, will use setup-ndk)
#   GOBIN - Go binary path (optional)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

DEBUG_MODE="${1:-}"

echo "=== sb-lxf Android AAR Build ==="
echo "Repository: $REPO_ROOT"
echo "Debug mode: $DEBUG_MODE"

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed or not in PATH"
    exit 1
fi

GO_VERSION=$(go version)
echo "Go version: $GO_VERSION"

# Check Android NDK
if [ -z "${ANDROID_NDK_HOME:-}" ]; then
    echo "ANDROID_NDK_HOME not set, attempting to find NDK..."
    # Try common locations
    if [ -d "/usr/local/share/android-ndk" ]; then
        export ANDROID_NDK_HOME="/usr/local/share/android-ndk"
    elif [ -d "$HOME/android-sdk/ndk/28.0.13004108" ]; then
        export ANDROID_NDK_HOME="$HOME/android-sdk/ndk/28.0.13004108"
    elif [ -d "$HOME/Library/Android/sdk/ndk/28.0.13004108" ]; then
        export ANDROID_NDK_HOME="$HOME/Library/Android/sdk/ndk/28.0.13004108"
    fi
fi

if [ -z "${ANDROID_NDK_HOME:-}" ] || [ ! -d "$ANDROID_NDK_HOME" ]; then
    echo "ERROR: Android NDK not found. Please set ANDROID_NDK_HOME or install NDK r28c"
    echo "Download: https://developer.android.com/ndk/downloads"
    exit 1
fi

echo "Android NDK: $ANDROID_NDK_HOME"
export PATH="$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin:$PATH"

# Check gomobile
if ! command -v gomobile &> /dev/null; then
    echo "Installing gomobile..."
    go install golang.org/x/mobile/cmd/gomobile@latest
    go install golang.org/x/mobile/cmd/gobind@latest
    gomobile init
fi

echo "gomobile: $(gomobile version 2>/dev/null || echo 'unknown')"

# Build tags for Android
BUILD_TAGS="with_quic,with_wireguard,with_utls,with_acme,with_clash_api,with_tailscale,with_ccm,with_ocm,with_cloudflared,with_usbip,with_openvpn,with_openconnect,badlinkname,tfogo_checklinkname0"

if [ "$DEBUG_MODE" = "debug" ]; then
    echo "Building DEBUG version..."
    BUILD_TAGS="$BUILD_TAGS,debug"
    LDFLAGS=""
else
    echo "Building RELEASE version..."
    LDFLAGS="-s -w -buildid="
fi

# Get version
VERSION=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
echo "Version: $VERSION"

# Build libbox AAR
echo ""
echo "Building libbox AAR..."
echo "Build tags: $BUILD_TAGS"
echo "LDFLAGS: -X 'github.com/sagernet/sing-box/constant.Version=$VERSION' $LDFLAGS"

go run ./cmd/internal/build_libbox -target android \
    -tags "$BUILD_TAGS" \
    -ldflags "-X 'github.com/sagernet/sing-box/constant.Version=$VERSION' $LDFLAGS"

# Check output
if [ -f "libbox.aar" ]; then
    AAR_SIZE=$(du -h libbox.aar | cut -f1)
    echo ""
    echo "✅ Build successful!"
    echo "Output: libbox.aar ($AAR_SIZE)"
    
    # Show contents
    echo ""
    echo "AAR contents:"
    unzip -l libbox.aar | grep "\.so" | head -10
else
    echo "ERROR: libbox.aar not found after build"
    exit 1
fi

echo ""
echo "=== Build complete ==="
echo "Output file: $(pwd)/libbox.aar"
echo "Next step: Copy to sb-AI app/libs/libbox.aar"

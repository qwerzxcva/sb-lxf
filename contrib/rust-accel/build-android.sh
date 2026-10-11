#!/usr/bin/env bash
# 构建 sb-lxf Rust 加速模块的 Android 静态库（aarch64）。
# 用法：
#   ./build-android.sh [NDK_ROOT]
# 输出：target/aarch64-linux-android/release/libsb_lxf_rust_accel.a
set -euo pipefail

NDK="${1:-${ANDROID_NDK:-$HOME/android-sdk/ndk/30.0.16248370}}"
PRE="$NDK/toolchains/llvm/prebuilt/linux-x86_64"
export CC_aarch64_linux_android="$PRE/bin/aarch64-linux-android26-clang"
export AR_aarch64_linux_android="$PRE/bin/llvm-ar"

cd "$(dirname "$0")"
echo "→ building sb_lxf_rust_accel for aarch64-linux-android (NDK: $NDK)"
cargo rustc --target aarch64-linux-android --release --lib
A="target/aarch64-linux-android/release/libsb_lxf_rust_accel.a"
ls -lh "$A"
echo "✅ 静态库就绪：$A（cgo LDFLAGS 用 -Ltarget/aarch64-linux-android/release -lsb_lxf_rust_accel）"
# 校验 FFI 符号
"$PRE/bin/llvm-nm" "$A" | grep -i "T sb_rust" || { echo "⚠ 未找到 sb_rust 导出符号"; exit 1; }
echo "✓ sb_rust_* FFI 符号已导出"

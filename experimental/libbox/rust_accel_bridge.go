//go:build with_rust_accel
// +build with_rust_accel

// 当启用 with_rust_accel 时，把 contrib/rust-accel（cgo 链接 Rust 静态库）
// 编译进 libbox。cgo_rust.go 的 LDFLAGS 会在链接阶段拉入
// libsb_lxf_rust_accel.a（需先用 build-android.sh 生成，或 CI 预构建）。
// 默认（未启用该标签）时本文件不参与编译，构建走 purego.go 回退，行为不变。
package libbox

import _ "github.com/sagernet/sing-box/contrib/rust-accel"

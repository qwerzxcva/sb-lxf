# sb-lxf Rust 加速模块（contrib/rust-accel）

零第三方依赖的 Rust 计算加速模块，通过 **C-ABI FFI** 暴露给 Go（cgo），
默认 **关闭**，启用构建标签 `with_rust_accel` 时链接。

## 提供的能力

| 能力 | Rust 实现 | 纯 Go 回退 |
|------|-----------|-----------|
| LPM / CIDR 最长前缀匹配（v4+v6） | `LpmTable`（每掩码层有序 HashMap，查询 O(层数)） | `net/netip` 线性扫描 |
| FNV-1a 64 位哈希 | `fnv1a_64` | `hash/fnv` |
| 轻量 64 位去重/负载哈希 | `quick_hash64` | 自实现 |

- Rust 版：静态分配、零 GC 压力、`opt-level=3 + fat LTO`，路由 IP 匹配
  热点（大量 CIDR 规则时）比 Go 线性扫描快，且无堆碎片。
- 纯 Go 版：零外部依赖，规则少时（<100 条）性能可接受，作为默认/降级。

## 目录

```
contrib/rust-accel/
├── Cargo.toml          # crate 定义（staticlib + rlib）
├── src/lib.rs          # LPM/哈希 实现 + C-ABI FFI 导出 + 单元测试
├── sb_rust_accel.h     # C 头（cgo include 用）
├── cgo_rust.go         # Go cgo 封装（//go:build with_rust_accel）
├── purego.go           # Go 纯实现（//go:build !with_rust_accel）
└── build-android.sh    # NDK 交叉编译 aarch64 静态库
```

两个 Go 文件暴露 **相同公开 API**（`rustaccel.NewLpmTable / InsertCIDR /
MatchV4 / MatchV6 / Fnv1a64 / QuickHash64 / Version / UsesRust`），
调用方无需改动，仅构建标签决定是否链接 Rust。

## 构建

```bash
# 本机单元测试（4 个用例）
cargo test

# Android aarch64 静态库（默认 NDK：~/android-sdk/ndk/30.0.16248370）
./build-android.sh
```

## 在 sb-lxf 中启用（可选）

1. 在 CI / 本地先生成静态库：`bash contrib/rust-accel/build-android.sh`
2. 给 `libbox` 的构建加 tag `with_rust_accel`：
   `make lib_android RUST_ACCEL=1`（build_libbox 检测到后会注入
   `-tags with_rust_accel` 并把静态库放好）。
3. `cgo_rust.go` 通过
   `#cgo LDFLAGS: -L${SRCDIR}/target/release -lsb_lxf_rust_accel -lm -lpthread`
   链接。

未启用时走 `purego.go`，默认构建（slim / full）行为完全不变。

## 为什么是 LPM / 哈希

- LPM（最长前缀匹配）是路由 IP-CIDR 规则的**经典热点**：每包都要在几十~几千
  条前缀里找最具体命中。Rust 的层式 HashMap 比 Go 线性扫描稳定更快。
- 哈希用于连接跟踪 / 去重键，纯计算、无 I/O，最适合做跨语言 FFI 隔离。
- 两者都**无状态 / 静态分配**，FFI 边界干净，不引入 Android 内存风险。

## 限制

- 目前 **未接入** sb-lxf 路由内核（`protocol/route`）作为加速器——
  本模块是独立的可选加速原语，后续可在 `route` 层引入 LPM 替换线性扫描。
- IPv6 解析为简化实现（标准 8 组 + 单 `::`，不支持 v4-mapped 混合）。
- 纯 Go 回退是线性扫描（O(规则数)），规则极多时应启用 Rust。

## 安全

- FFI 边界用 `catch_unwind` 包裹，Rust 侧 panic 不会穿透到 Go。
- 指针参数均做 null 检查；`sb_rust_free_cstr` 管理 C 字符串生命周期。

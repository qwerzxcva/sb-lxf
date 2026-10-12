# sb-lxf (tsbAI) Changelog

Based on SagerNet/sing-box **v1.15.0-alpha.12**.

## v1.15.0-ts.2 (2026-10-11)
- Ported Leadaxe (sing-box-lx) Android / low-power extensions onto the
  v1.15.0-alpha.12 base so sb-AI runs a single unified kernel:
  - balancer pool for `urltest` `round_robin` (`pool_size`, `pool_tolerance`,
    `sticky_hash`) — replaces the failover/penalty/tick split.
  - XHTTP (Xray splithttp) v2ray transport, XMUX + padding, always compiled
    in sb-lxf (de-gated from `with_xhttp`).
  - Xray gRPC `service_name` custom-path (SPEC 093) + lite gRPC variant.
  - outbound `chain` (SPEC 073) behind `with_lx_chain` (default off).
  - libbox `CommandDNS` structured DNS-query event stream (SPEC 018).
  - desktop `boxdd` daemon gated behind `with_boxdd` (Android default skips it).
  - `group.Now()` on `OutboundGroup` (Selector + URLTest) for current-node
    diagnostics used by the chain/observability paths.
- The Leadaxe OOM-profile builder (libbox `internal/oomprofile`) links via
  `//go:linkname` into `runtime/pprof`; Go 1.23+ requires
  `-ldflags=-checklinkname=0` to link the libbox test/executable. The
  default `./cmd/sing-box` and the Android `.so` do not link oomprofile, so
  they build clean. CI/workflow sets the flag where the libbox binary links.
- WireGuard: default build uses upstream `sagernet/wireguard-go` (no fork).
  The Leadaxe AWG2/AWG3 + lazy/suspend/idle extensions are kept only behind
  their original `with_awg*` build tags (they need the Leadaxe wireguard-go
  fork); the standard kernel does not enable them.
- Removed DNS-transport-group (Leadaxe-only) — sb-AI doesn't use it.
- Removed broken Leadaxe-specific regression tests that target 1.14 API
  surfaces not present in the 1.15 base.

## v1.15.0-ts.1 (2026-10-11)
- Upstream Go TUN stack exposed as `tsbAI` (alias `go`).
- tsbAI supports GSO/splice, multi-queue, memory-pressure OOM-killer.
- Kept upstream sing-tun pinned (includes the Go stack).

## Build notes
- Default: `go build ./cmd/sing-box` (no tags) — Android libbox path.
- `-tags with_lx_chain` adds the outbound `chain` group.
- `-tags with_boxdd` adds the desktop daemon.
- `-tags with_awg2 with_awg3` (needs Leadaxe wireguard-go fork) for AWG.
- Link libbox/oomprofile on Go 1.23+: `-ldflags=-checklinkname=0`.
## v1.15.0-ts.3 (2026-10-11)
- CI: use `android-actions/setup-android@v3` (real SDK + NDK + platforms) instead of
  fake SDK. Install `android-21`/`android-24` platforms for `gomobile bind`.
- CI: install `github.com/sagernet/gomobile` fork (matches go.mod, not x/mobile).
- Fix: `transport/v2rayxhttp/http3.go`, `http3_stub.go`, `http_version.go` restored
  from Leadaxe lx-1.14.3 (batch-patch corruption).
- Fix: Reality `KeyShare` constants + `OutboundRealityOptions.KeyShare` field
  (Leadaxe SPEC 089).
- Fix: `utls.HelloChrome_155`/`HelloFirefox_148` mapped to `133`/`120`
  (metacubex/utls v1.8.7 max is 133/120).

## v1.15.0-ts.4 (2026-10-12) — Android-only power/size optimization
- Add `-slim` flag to `cmd/internal/build_libbox` + `make lib_android_slim`
  target: Android-first build profile for sb-AI release builds.
- Slim profile drops only the transports sb-AI has no reference to:
  `with_usbip`, `with_openvpn`, `with_openconnect`.
- `with_tailscale` + all `ts_omit_*` are KEPT in slim because sb-AI exposes
  "tailscale" as a first-class endpoint type (NodeEditorScreen /
  SingBoxConfigGenerator ENDPOINT_TYPES). Removing it would break that feature.
- Cross-fork validation: matches Karing (KaringX/sing-box) and PuerNya Android
  AAR builds (same drop-usbip/openvpn/openconnect philosophy). sb-lxf's extra
  power edge is the `tsbAI` TUN stack (GSO / multi_queue / oom-killer) vs the
  gvisor stack those forks use.
- Reference doc: `release/ANDROID_BUILD_TAGS.md` documents both profiles.
- The default `lib_android` (full profile) is kept intact; existing CI uses it.
- Rebased local onto latest remote (synced wireguard/utls sync from lx-1.14.3).

## v1.15.0-ts.5 (2026-10-12) — attempt to enable lx_idle_suspend (SPEC 020) in slim build (deferred)
- Add `with_lx_idle_suspend` to the Android slim profile (`lib_android_slim`).
  This is the Leadaxe idle-suspend optimization (SPEC 020): idle WG/Tailscale
  endpoints are suspended after a configurable idle window, freeing
  recv-worker bufsArrs (the dominant GC-scan holder) and preventing background
  work for unused nodes.
- Reference: enabled in FlClash, Bettbox, sing-box-lx, Karing, PuerNya
  Android-optimized builds. Idle tunnels do zero work when unused.
- The full (`lib_android`) build remains unchanged; slim-only.
## v1.15.0-ts.6 (2026-10-12) — slim release + idle-suspend active
- Build tag: `with_lx_idle_suspend` (SPEC 020) is now in the slim profile
  (`lib_android_slim`). Idle WireGuard/Tailscale endpoints suspend after the
  configured window, freeing recv-worker bufsArrs (dominant GC-scan holder) and
  cutting background work for unused tunnels to zero.
- Slim profile summary:
  ```
  quic, wireguard, utls, naive_outbound, clash_api,
  tailscale(+all ts_omit), lx_idle_suspend, badlinkname, tfogo_checklinkname0
  ```
  Dropped: usbip, openvpn, openconnect (none referenced by sb-AI).
- Cross-fork validation vs Karing / PuerNya / shtorm-7 / reF1nd / Kwisma:
  the slim subset matches the minimal set used by all of them; sb-lxf adds
  tsbAI (pure-Go TUN with splice/GSO/multi_queue/oom-killer) which is a power
  advantage over their gvisor stack.
- CI: the `build-android-aar-slim` job builds and uploads `libbox-{ver}-slim.aar`
  for all pushes/releases. This AAR is ready for sb-AI consumption.
## v1.15.0-ts.7 — optional Rust LPM / hash accelerator（新增）
- 新增 `contrib/rust-accel/`：零第三方依赖的 Rust 加速模块，通过 C-ABI FFI
  暴露给 Go（cgo），默认关闭，启用构建标签 `with_rust_accel` 时链接。
  - `LpmTable`（LPM / CIDR 最长前缀匹配，v4+v6）：每掩码层有序 HashMap，
    查询 O(层数)，静态分配、零 GC 压力；适用于大量 IP-CIDR 规则的路由热点。
  - `fnv1a_64` / `quick_hash64`：连接跟踪 / 去重键哈希。
  - 两个 Go 文件 (`cgo_rust.go` / `purego.go`) 暴露同一公开 API；未启用时
    走纯 Go 回退（线性扫描），默认构建行为完全不变。
  - CI / Makefile：`make rust_accel` 构建 aarch64 静态库（~2 MB 压缩后）；
    `make lib_android_slim_rust` 生成包含 Rust 加速的 slim AAR。
- 入口桥接：`experimental/libbox/rust_accel_bridge.go`（//go:build with_rust_accel）。
  调用方添加 `import _ "github.com/sagernet/sing-box/contrib/rust-accel"` 即可生效。
- 安全：FFI 边界用 `catch_unwind` 包裹，指针 null 检查全覆盖；Rust panic
  不会穿透到 Go；C 字符串通过 `sb_rust_free_cstr` 管理生命周期。

## v1.15.0-ts.8 (2026-10-12) — Android-only slim：协议族构建标签化 + 裁掉 Linux 桌面构建
- 将 6 个 sb-AI 永不使用、但默认常驻编译的协议族（anytls / shadowtls /
  snell / tor / ssh / masque）改为构建标签可选：新增 `with_anytls`、
  `with_shadowtls`、`with_snell`、`with_tor`、`with_ssh`、`with_masque`。
  - 注册逻辑移入 `include/proto_<name>.go`（`//go:build with_<name>`），
    配 `include/proto_<name>_stub.go`（`//go:build !with_<name>`）空实现，
    沿用 openvpn/openconnect 的 stub 模式。`include/quic.go` 里的 masque
    出站也改为调 `registerMasqueOutbound`（去掉了直接 import）。
  - 默认 / full 构建把这 6 个标签加进 `sharedTags`，行为不变（仍全量编译）。
  - `lib_android_slim`（sb-AI 发布档）把 6 个标签全部排除：sb-AI 配置生成器
    只发 vless/vmess/trojan/shadowsocks/socks/hysteria2/hysteria/tuic/
    wireguard/tailscale，丢弃死协议 → AAR 再小 ~3MB、运行期无这些协议的
    闲置 handler。
- CI `build-tsbai.yml`：删掉 `build-linux-amd64` 任务与 release 里的
  Linux 二进制下载（纯 Android 内核，不再产桌面 binary；编译更快、
  发布产物更干净）。
- sb-AI 侧：`scripts/fetch-libbox.sh` 默认档 → `v1.15.0-ts.8`；
  `.github/workflows/build.yml` + `release.yml` 的 `LIBBOX_VERSION` → ts.8。

## v1.15.0-ts.9 (2026-10-13) — 修正 Android-only 裁剪 + CI 编译缓存
- **修正 ts.8 过度裁剪**：sb-AI 的 `OUTBOUND_TYPES`（NodeEditorScreen.kt:60）
  实际包含 `tor` 和 `ssh`，ts.8 误删了它们。现恢复 slim 里的
  `with_tor` / `with_ssh`（sb-AI 完美适配优先）。
- **新增裁剪** `with_naive_outbound`：sb-AI 节点类型里没有 naive
  （HTTP CONNECT 隧道），slim 档不再编译 naive 出站构造器。
  （naive 入站仍无条件注册，彻底裁需再门控入站注册，留待后续。）
- **CI 编译效率**：`build-tsbai.yml` 两个 Android AAR 任务新增
  `actions/cache@v4` 缓存 `~/go/pkg/mod`（模块缓存，按 go.mod/go.sum
  hash）+ `~/.cache/go-build`（构建缓存，按 Go 版本）。增量重建跳过
  依赖下载与共享包重编译，显著加快每次 push 的编译速度。
- 净效果（slim）：裁掉 anytls / shadowtls / snell / masque /
  naive_outbound 共 5 族，保留 tor / ssh；AAR 更小、运行期无这些协议的
  闲置 handler；tor/ssh 节点类型在 sb-AI 里仍可正常生成。

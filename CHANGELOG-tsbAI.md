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
- Add `-slim` flag to `cmd/internal/build_libbox`: drops `with_usbip`,
  `with_openvpn`, `with_openconnect`, `with_tailscale` + all `ts_omit_*` from
  the Android AAR build. sb-AI consumes quic/wireguard/utls/naive/clash_api
  only — the dropped features add tens of MB and background-service overhead
  with no benefit on Android.
- New Makefile target `make lib_android_slim` (replaces `make lib_android` as
  the recommended release path for Android-only clients).
- Reference doc: `release/ANDROID_BUILD_TAGS.md` documents both profiles.
- The default `lib_android` (full profile) is kept intact; existing CI uses it.
- Rebased local onto latest remote (synced wireguard/utls sync from lx-1.14.3).

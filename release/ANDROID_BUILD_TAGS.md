# sb-lxf Android build profiles

Two Android AAR build profiles exist (both invoke `make lib_android*`, which
calls `cmd/internal/build_libbox -target android`).

## `lib_android` (full)
Tags:
```
with_quic,with_wireguard,with_utls,with_naive_outbound,with_clash_api,with_usbip,
with_openvpn,with_openconnect,badlinkname,tfogo_checklinkname0,
with_tailscale,ts_omit_*(logtail,ssh,drive,taildrop,webclient,doctor,capture,kube,aws,synology,bird)
```
- Full feature set (usbip / openvpn / openconnect / tailscale enabled).
- Use when you need those transports on Android.

## `lib_android_slim` (recommended for sb-AI release builds)
`make lib_android_slim` = `build_libbox -target android -slim`.
Drops: `with_usbip`, `with_openvpn`, `with_openconnect` (unreferenced by sb-AI).
`with_tailscale` + all `ts_omit_*` are KEPT because sb-AI exposes Tailscale as a
first-class endpoint type. Adds: `with_lx_idle_suspend` (SPEC 020, idle-suspend).

### Android-only protocol cut (ts.8)
The slim profile now also gates 6 protocol families behind new build tags and
excludes them from sb-AI (sb-AI's config generator only emits
vless/vmess/trojan/shadowsocks/socks/hysteria2/hysteria/tuic/wireguard/tailscale):

| tag | protocol | reason |
|-----|----------|--------|
| `with_anytls` | anytls | not used by sb-AI |
| `with_shadowtls` | shadowtls | not used by sb-AI |
| `with_snell` | snell | not used by sb-AI |
| `with_tor` | tor | not used by sb-AI |
| `with_ssh` | ssh | not used by sb-AI |
| `with_masque` | masque (WARP) | not used by sb-AI |

Default/full build keeps all 6 (tags added to `sharedTags`); only the slim
profile excludes them. The registration moves to `include/proto_<name>.go`
(`//go:build with_<name>`) with a no-op stub at `include/proto_<name>_stub.go`
(`//go:build !with_<name>`), mirroring the openvpn/openconnect pattern.

Benefits:
- Smaller AAR + lower runtime CPU/battery: unused transports (usbip / openvpn /
  openconnect) are not linked, so no idle background handlers for them.
- Idle-suspend (`with_lx_idle_suspend`, SPEC 020): idle WireGuard/Tailscale
  endpoints are suspended after an idle window, releasing their recv-worker
  buffers (the dominant GC-scan holder). Zero background work when a tunnel
  is idle.
- Tailscale is kept so sb-AI's Tailscale endpoint type keeps working.
- Android-only: the desktop CLI (`cmd/sing-box`) and Apple/desktop helpers are
  never compiled into the AAR regardless of profile.

Cross-fork validation: this slim profile matches Karing (KaringX/sing-box) and
PuerNya's Android AAR builds, which also drop usbip/openvpn/openconnect while
keeping quic/wireguard/utls/naive/clash_api (+tailscale in Karing). sb-lxf's
additional power edge is the `tsbAI` TUN stack (GSO / multi_queue / oom-killer)
vs the gvisor stack those forks use.

The kernel is Android-first: `make lib_android_slim` is the release target.

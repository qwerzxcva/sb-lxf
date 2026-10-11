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
Drops: `with_usbip`, `with_openvpn`, `with_openconnect`, `with_tailscale` + all
`ts_omit_*`. Keeps the core proxy stack sb-AI actually consumes:
`with_quic, with_wireguard, with_utls, with_naive_outbound, with_clash_api,
badlinkname, tfogo_checklinkname0`.

Benefits:
- Smaller AAR (no tailscale / usbip / openvpn / openconnect code linked in).
- Lower runtime CPU/battery (no unused background services or protocol handlers
  that are never referenced on Android).
- Android-only: the desktop CLI (`cmd/sing-box`) and Apple/desktop helpers are
  never compiled into the AAR regardless of profile.

The kernel is Android-first: `make lib_android_slim` is the release target.

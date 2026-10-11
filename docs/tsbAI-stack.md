# tsbAI — Go TUN stack (renamed from "go")

Based on SagerNet/sing-box v1.15.0-alpha.12. The Go TUN stack (new in 1.15)
is exposed as `stack: "tsbAI"` in addition to the `"go"` alias.

## Features

- **GSO / splice** — outbound TLS spliced directly into the socket when
  TLS record splitting is disabled.
- **Multi-queue** (`multi_queue: true`) — per-queue engine, Linux only.
- **Memory pressure + OOM killer** — idle conn release under memory
  pressure, coordinated with the oom-killer service.

## MTU

`tsbAI` + MTU 9000 is the recommended Android combo. GSO auto-enables when
MTU < 49152.

## How to use

```
"tun": {
  "stack": "tsbAI",      // or "go"
  "mtu": 9000,
  "multi_queue": true,   // optional, Linux only
  "enable_ipoib": false
}
```

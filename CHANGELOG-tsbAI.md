# sb-lxf (tsbAI) Changelog

Based on SagerNet/sing-box **v1.15.0-alpha.12**.

## v1.15.0-ts.1 (2026-10-11)
- Upstream Go TUN stack exposed as `tsbAI` (alias `go`).
- tsbAI supports GSO/splice, multi-queue, memory-pressure OOM-killer.
- Kept upstream sing-tun pinned (includes the Go stack).

## Not yet done
- Leadaxe-specific extensions (balancer pool, AWG2, XHTTP XMUX range,
  utls-firefox148, sing-mux 256) are NOT in this base — those require
  Leadaxe's 1.14.3 fork. sb-AI should use this kernel for the tsbAI TUN
  stack; features exclusive to Leadaxe-lx are optional.
- CI Android AAR build is provided in .github/workflows/ (build-tsbai.yml).

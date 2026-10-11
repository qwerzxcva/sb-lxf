#!/usr/bin/env bash
# 下载预编译 sb-lxf (tsbAI) 内核 AAR
# 基于 sing-box-lx 源码，新增 go TUN stack 支持
set -euo pipefail

LIBBOX_VERSION="${LIBBOX_VERSION:-v1.15.0-tsbai.1}"
REPO="QWERZXCVA/sb-lxf"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$ROOT/app/libs/libbox.aar"
AAR_NAME="libbox-${LIBBOX_VERSION#v}.aar"
BASE_URL="https://github.com/$REPO/releases/download/$LIBBOX_VERSION"
MARKER="$ROOT/app/libs/.libbox.version"

mkdir -p "$(dirname "$DEST")"

if [[ -f "$DEST" && -f "$MARKER" && "$(cat "$MARKER")" == "$LIBBOX_VERSION" && "${FORCE:-0}" != "1" ]]; then
  echo "✓ libbox.aar already $LIBBOX_VERSION — skipping"
  exit 0
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "→ Fetching $AAR_NAME ($REPO @ $LIBBOX_VERSION)"
curl -fsSL --retry 3 -o "$TMP/$AAR_NAME" "$BASE_URL/$AAR_NAME"
curl -fsSL --retry 3 -o "$TMP/SHA256SUMS" "$BASE_URL/SHA256SUMS"

(
  cd "$TMP"
  grep -E "  ?$AAR_NAME$" SHA256SUMS > expected.sum
  sha256sum -c expected.sum
)

python3 - "$TMP/$AAR_NAME" "$DEST" <<'PY'
import sys, zipfile
src, dst = sys.argv[1], sys.argv[2]
zin = zipfile.ZipFile(src)
keep = [n for n in zin.namelist()
        if not n.startswith('jni/') or n.startswith('jni/arm64-v8a/')]
with zipfile.ZipFile(dst, 'w', zipfile.ZIP_STORED) as zout:
    for n in keep:
        zi = zipfile.ZipInfo(n, date_time=(2026, 1, 1, 0, 0, 0))
        zi.external_attr = 0o644 << 16
        zout.writestr(zi, zin.read(n))
abis = sorted({n.split('/')[1] for n in keep if n.startswith('jni/')})
print(f"✓ kept ABIs: {abis}")
assert abis == ['arm64-v8a'], f"expected arm64-v8a only, got {abis}"
PY

printf '%s' "$LIBBOX_VERSION" > "$MARKER"
echo "✅ $DEST ← $LIBBOX_VERSION ($(du -h "$DEST" | cut -f1))"

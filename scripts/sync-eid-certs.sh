#!/usr/bin/env bash
# Sync Belgian eID CA certificates into go/internal/eid/certs/ (Web eID trust store).
#
# Sources:
#   1) http://certs.eid.belgium.be/     — directory index (roots, citizen*, foreigner*)
#   2) http://crt.eidpki.belgium.be/eid/ — AIA (no index): eidf*, eidc*, brca* → belgiumrca*
#
# Usage (repo root):
#   bash scripts/sync-eid-certs.sh
#   bash scripts/sync-eid-certs.sh --dry-run
#
# Requires: curl, openssl, python3
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CERTS_DIR="${ROOT}/go/internal/eid/certs"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

DRY_RUN=0
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=1

UA="petsFollow-eid-sync/1.0"
MIRROR="http://certs.eid.belgium.be"
AIA="http://crt.eidpki.belgium.be/eid"

added=0
skipped=0
failed=0

log() { printf '%s\n' "$*" >&2; }

is_x509() {
  local f="$1"
  [[ -s "$f" ]] || return 1
  if head -c 200 "$f" | grep -qiE '<html|<!doctype|not found|404'; then
    return 1
  fi
  openssl x509 -inform PEM -in "$f" -noout -subject >/dev/null 2>&1 && return 0
  openssl x509 -inform DER -in "$f" -noout -subject >/dev/null 2>&1 && return 0
  return 1
}

to_pem() {
  local src="$1" dest="$2"
  if openssl x509 -inform PEM -in "$src" -out "$dest" 2>/dev/null; then
    return 0
  fi
  openssl x509 -inform DER -in "$src" -out "$dest"
}

fp_der() {
  openssl x509 -in "$1" -outform DER 2>/dev/null | sha256sum | awk '{print $1}'
}

# Save verified cert as PEM under CERTS_DIR/$dest_name (.crt).
save_cert() {
  local dest_name="$1" src="$2"
  dest_name="${dest_name%.crt}.crt"
  local dest="${CERTS_DIR}/${dest_name}"
  local pem="${TMP}/norm.${dest_name}"
  if ! to_pem "$src" "$pem"; then
    log "WARN: normalize failed for ${dest_name}"
    failed=$((failed + 1))
    return 1
  fi
  local new_fp
  new_fp="$(fp_der "$pem")"
  if [[ -f "$dest" ]]; then
    local old_fp
    old_fp="$(fp_der "$dest")"
    if [[ "$old_fp" == "$new_fp" ]]; then
      # Same cert: rewrite DER → PEM if needed (loader accepts both; pack stays PEM).
      if ! openssl x509 -inform PEM -in "$dest" -noout >/dev/null 2>&1; then
        if [[ "$DRY_RUN" -eq 1 ]]; then
          log "DRY-RUN PEM-REWRITE ${dest_name}"
        else
          cp "$pem" "$dest"
          log "PEM-REWRITE ${dest_name}"
        fi
      fi
      skipped=$((skipped + 1))
      return 0
    fi
    log "WARN: collision different cert — keeping existing ${dest_name}"
    skipped=$((skipped + 1))
    return 0
  fi
  if [[ "$DRY_RUN" -eq 1 ]]; then
    log "DRY-RUN ADD ${dest_name}"
    added=$((added + 1))
    return 0
  fi
  cp "$pem" "$dest"
  log "ADD ${dest_name}"
  added=$((added + 1))
}

download() {
  curl -fsSL -A "$UA" --connect-timeout 8 --max-time 25 -o "$2" "$1" 2>/dev/null
}

keep_mirror_name() {
  case "$1" in
    belgiumrca*|belgiumrs*|citizen*|foreigner*|brca*) return 0 ;;
    *) return 1 ;;
  esac
}

# --- 1) Mirror index ---
log "==> Index ${MIRROR}/"
curl -fsSL -A "$UA" --max-time 60 -o "${TMP}/index.html" "${MIRROR}/"
mapfile -t LINKS < <(grep -oE 'href="[^"]+\.(crt|cer|pem|der)"' "${TMP}/index.html" \
  | sed -E 's/^href="//; s/"$//; s|^\./||; s|^https?://certs\.eid\.belgium\.be/||' \
  | sort -u)
log "Mirror links: ${#LINKS[@]}"

for rel in "${LINKS[@]}"; do
  base="$(basename "$rel")"
  keep_mirror_name "$base" || continue
  out="${TMP}/m.${base}"
  if download "${MIRROR}/${rel}" "$out" && is_x509 "$out"; then
    dest="${base%.*}.crt"
    if [[ "$base" == brca.crt ]]; then dest="belgiumrca.crt"
    elif [[ "$base" =~ ^brca([0-9]+)\.crt$ ]]; then dest="belgiumrca${BASH_REMATCH[1]}.crt"
    fi
    save_cert "$dest" "$out" || true
  else
    failed=$((failed + 1))
  fi
done

# --- 2) AIA probes (no directory listing) ---
log "==> Probe ${AIA}/"
hits_eidf=0 hits_eidc=0 hits_eida=0 hits_citizen=0 hits_foreigner=0 hits_brca=0

probe_aia() {
  local server_name="$1" dest_name="$2"
  local out="${TMP}/a.${server_name}"
  if download "${AIA}/${server_name}" "$out" && is_x509 "$out"; then
    save_cert "$dest_name" "$out" || true
    return 0
  fi
  return 1
}

for n in "" 2 3 4 5 6 7 8 9 10; do
  if probe_aia "brca${n}.crt" "belgiumrca${n}.crt"; then hits_brca=$((hits_brca + 1)); fi
  if probe_aia "belgiumrca${n}.crt" "belgiumrca${n}.crt"; then hits_brca=$((hits_brca + 1)); fi
  probe_aia "belgiumrs${n}.crt" "belgiumrs${n}.crt" || true
done

for year in $(seq 2018 2026); do
  for mon in $(seq 1 12); do
    mm=$(printf '%02d' "$mon")
    yymm="${year}${mm}"
    if probe_aia "eidf${yymm}.crt" "eidf${yymm}.crt"; then hits_eidf=$((hits_eidf + 1)); fi
    if probe_aia "eidc${yymm}.crt" "eidc${yymm}.crt"; then hits_eidc=$((hits_eidc + 1)); fi
    if probe_aia "eida${yymm}.crt" "eida${yymm}.crt"; then hits_eida=$((hits_eida + 1)); fi
    if probe_aia "citizen${yymm}.crt" "citizen${yymm}.crt"; then hits_citizen=$((hits_citizen + 1)); fi
    if probe_aia "foreigner${yymm}.crt" "foreigner${yymm}.crt"; then hits_foreigner=$((hits_foreigner + 1)); fi
  done
  log "  year ${year}: eidf=${hits_eidf} eidc=${hits_eidc}"
done

log "AIA hits: brca=${hits_brca} eidf=${hits_eidf} eidc=${hits_eidc} eida=${hits_eida} citizen=${hits_citizen} foreigner=${hits_foreigner}"
[[ "$hits_eida" -eq 0 ]] && log "404 pattern: ${AIA}/eidaYYYYMM.crt"
[[ "$hits_foreigner" -eq 0 ]] && log "404 pattern: ${AIA}/foreignerYYYYMM.crt (use mirror foreigner* + AIA eidf*)"

# --- 3) Dedup by SHA256(DER) — prefer belgiumrca > belgiumrs, eidf > foreigner ---
if [[ "$DRY_RUN" -eq 0 ]]; then
  log "==> Deduplicate"
  python3 - "$CERTS_DIR" <<'PY'
import hashlib, subprocess, sys
from pathlib import Path
certs = Path(sys.argv[1])

def der(p: Path) -> bytes:
    for inform in ("PEM", "DER"):
        r = subprocess.run(
            ["openssl", "x509", "-inform", inform, "-in", str(p), "-outform", "DER"],
            capture_output=True,
        )
        if r.returncode == 0 and r.stdout:
            return r.stdout
    raise RuntimeError(p)

def score(name: str):
    pref = 0
    if name.startswith("belgiumrca"): pref = 100
    elif name.startswith("belgiumrs"): pref = 90
    elif name.startswith("eidf"): pref = 80
    elif name.startswith("eidc"): pref = 70
    elif name.startswith("citizen"): pref = 60
    elif name.startswith("foreigner"): pref = 50
    return (pref, -len(name), name)

by_fp = {}
for f in sorted(certs.glob("*.crt")):
    by_fp.setdefault(hashlib.sha256(der(f)).hexdigest(), []).append(f.name)

removed = 0
for names in by_fp.values():
    if len(names) < 2:
        continue
    keep = sorted(names, key=score, reverse=True)[0]
    for n in names:
        if n == keep:
            continue
        print(f"DEDUP remove {n} (same as {keep})", file=sys.stderr)
        (certs / n).unlink(missing_ok=True)
        removed += 1
print(f"dedup_removed={removed}", file=sys.stderr)
PY

  log "==> Normalize remaining DER → PEM"
  python3 - "$CERTS_DIR" <<'PY'
import subprocess, sys
from pathlib import Path
certs = Path(sys.argv[1])
n = 0
for f in sorted(certs.glob("*.crt")):
    pem_ok = subprocess.run(
        ["openssl", "x509", "-inform", "PEM", "-in", str(f), "-noout"],
        capture_output=True,
    ).returncode == 0
    if pem_ok:
        continue
    r = subprocess.run(
        ["openssl", "x509", "-inform", "DER", "-in", str(f), "-outform", "PEM"],
        capture_output=True,
    )
    if r.returncode != 0 or not r.stdout:
        print(f"WARN: cannot PEM-normalize {f.name}", file=sys.stderr)
        continue
    f.write_bytes(r.stdout)
    print(f"PEM-NORMALIZE {f.name}", file=sys.stderr)
    n += 1
print(f"pem_normalized={n}", file=sys.stderr)
PY
fi

log "==> Summary added=${added} skipped=${skipped} failed=${failed} files=$(find "$CERTS_DIR" -maxdepth 1 -name '*.crt' | wc -l)"
log "Update go/internal/eid/certs/README.md sync date after a full run."

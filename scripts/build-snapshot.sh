#!/usr/bin/env bash
set -euo pipefail

# Minimal cross-build helper for Linux/macOS (amd64/arm64).
# Produces tarballs and a SHASUMS256.txt file under dist/.

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
DIST_DIR="$ROOT_DIR/dist"
BIN_NAME="crush"
VERSION="${VERSION:-0.0.0-snapshot}"

mkdir -p "$DIST_DIR"

build_one() {
  local os="$1" arch="$2"
  local outdir="$DIST_DIR/${BIN_NAME}_${os}_${arch}"
  local outfile="$outdir/$BIN_NAME"
  echo "==> Building $os/$arch"
  mkdir -p "$outdir"
  env CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -ldflags "-s -w -X github.com/charmbracelet/crush/internal/version.Version=${VERSION}" \
    -o "$outfile" .
  (cd "$DIST_DIR" && tar -czf "${BIN_NAME}_${os}_${arch}.tar.gz" "${BIN_NAME}_${os}_${arch}")
  rm -rf "$outdir"
}

targets=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
)

for t in "${targets[@]}"; do
  build_one $t
done

echo "==> Writing checksums"
(cd "$DIST_DIR" && sha256sum ${BIN_NAME}_*.tar.gz > SHASUMS256.txt)
echo "Done. Artifacts in $DIST_DIR"

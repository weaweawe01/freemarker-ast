#!/usr/bin/env bash

set -u

PHASE="${1:-all}"

if [[ "$PHASE" != "all" && "$PHASE" != "basic" && "$PHASE" != "complex" && "$PHASE" != "nested" ]]; then
  echo "usage: $0 [all|basic|complex|nested]" >&2
  exit 2
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FM_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CASES_DIR="$SCRIPT_DIR/cases"
JAVA_OUT_DIR="$SCRIPT_DIR/out/java"
GO_OUT_DIR="$SCRIPT_DIR/out/go"
DIFF_DIR="$SCRIPT_DIR/diff"
MVN_BIN="${MVN_BIN:-mvn}"

if ! command -v go >/dev/null 2>&1; then
  echo "[ERROR] go not found in PATH" >&2
  exit 2
fi

if [[ "$MVN_BIN" = */* ]]; then
  if [[ ! -x "$MVN_BIN" ]]; then
    echo "[ERROR] maven executable not found: $MVN_BIN" >&2
    exit 2
  fi
else
  if ! command -v "$MVN_BIN" >/dev/null 2>&1; then
    echo "[ERROR] mvn not found in PATH" >&2
    exit 2
  fi
fi

if ! "$MVN_BIN" -v >/dev/null 2>&1; then
  echo "[ERROR] failed to run maven executable: $MVN_BIN" >&2
  exit 2
fi

mkdir -p "$JAVA_OUT_DIR" "$GO_OUT_DIR" "$DIFF_DIR"

echo "[INFO] compiling Java helper classes..."
if ! (cd "$FM_ROOT/freemarker-java" && "$MVN_BIN" -q -DskipTests compile); then
  echo "[ERROR] Java compile failed" >&2
  exit 1
fi

if [[ "$PHASE" == "all" ]]; then
  CASE_ROOT="$CASES_DIR"
  JAVA_OUT_ROOT="$JAVA_OUT_DIR"
  GO_OUT_ROOT="$GO_OUT_DIR"
  DIFF_ROOT="$DIFF_DIR"
else
  CASE_ROOT="$CASES_DIR/$PHASE"
  JAVA_OUT_ROOT="$JAVA_OUT_DIR/$PHASE"
  GO_OUT_ROOT="$GO_OUT_DIR/$PHASE"
  DIFF_ROOT="$DIFF_DIR/$PHASE"
fi

if [[ ! -d "$CASE_ROOT" ]]; then
  echo "[ERROR] case root not found: $CASE_ROOT" >&2
  exit 2
fi

mkdir -p "$JAVA_OUT_ROOT" "$GO_OUT_ROOT" "$DIFF_ROOT"
find "$JAVA_OUT_ROOT" -type f -name '*.ast' -delete 2>/dev/null || true
find "$GO_OUT_ROOT" -type f -name '*.ast' -delete 2>/dev/null || true
find "$DIFF_ROOT" -type f -name '*.diff' -delete 2>/dev/null || true

echo "[INFO] Java batch dump..."
if ! (
  cd "$FM_ROOT/freemarker-java" && \
  "$MVN_BIN" -q \
    -Dexec.mainClass=io.freemarker.astdump.BatchFileAstMain \
    -Dexec.args="\"$CASE_ROOT\" \"$JAVA_OUT_ROOT\"" \
    exec:java
); then
  echo "[ERROR] Java batch AST dump failed" >&2
  exit 1
fi

echo "[INFO] Go batch dump..."
if ! (
  cd "$FM_ROOT" && \
  GOCACHE="${GOCACHE:-/tmp/go-build}" go run ./cmd/fm-ast-dump-dir \
    --in-root "$CASE_ROOT" \
    --out-root "$GO_OUT_ROOT"
); then
  echo "[ERROR] Go batch AST dump failed" >&2
  exit 1
fi

echo "[INFO] Batch compare..."
if ! (
  cd "$FM_ROOT" && \
  GOCACHE="${GOCACHE:-/tmp/go-build}" go run ./cmd/fm-ast-compare-dir \
    --cases-root "$CASE_ROOT" \
    --oracle-root "$JAVA_OUT_ROOT" \
    --actual-root "$GO_OUT_ROOT" \
    --diff-root "$DIFF_ROOT"
); then
  exit 1
fi

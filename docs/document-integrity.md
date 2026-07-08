# Document Integrity & Link Validation

> 参考自 `AGENTS.md` 的 Document Integrity & Link Validation 节。每次文档变更后自动执行 3 层完整性检查。

## Layer 1 — Link Integrity Check

```bash
# macOS-compatible link integrity scan
CHANGED_FILE="$1"  # pass file path as argument, or set before running
echo "=== Link Integrity Check ==="
grep -oE '\[[^]]+\]\([^)]+\)' "$CHANGED_FILE" 2>/dev/null | \
  sed 's/\[[^]]*\]//' | tr -d '()' | while IFS= read -r target; do
  case "$target" in
    http*) echo "⏭️ EXTERNAL: $target" ;;
    #*) echo "⏭️ ANCHOR: $target" ;;
    *)  dir="$(dirname "$CHANGED_FILE")"
        if [ -f "$dir/$target" ] || [ -f "$target" ]; then
          echo "✅ $target"
        else
          echo "❌ MISSING: $target (referenced from $CHANGED_FILE)"
        fi ;;
  esac
done
```

## Layer 2 — Cross-Reference Symmetry

```bash
# Cross-reference symmetry check (extracts only from Markdown links `[...](...)`)
echo "=== Cross-Reference Symmetry ==="
grep -oE '\[[^]]+\]\(docs/[^)]+\)' AGENTS.md 2>/dev/null | \
  sed 's/.*(docs/ds/' | tr -d ')' | while IFS= read -r href; do
  if grep -q 'AGENTS.md' "$href" 2>/dev/null; then
    echo "✅ $href has backlink to AGENTS.md"
  else
    echo "⚠️  $href missing backlink to AGENTS.md"
  fi
done
```

## Layer 3 — Content Dedup Check

```bash
# Deduplication scan
CHANGED_FILE="$1"
[ -z "$CHANGED_FILE" ] && { echo "Usage: CHANGED_FILE=<path> $0"; exit 1; }
CHANGED="$(basename "$CHANGED_FILE")"
SKILL_DIR="$(dirname "$CHANGED_FILE")"
for peer_file in "$SKILL_DIR"/*.md "$SKILL_DIR"/references/*.md 2>/dev/null; do
  [ "$peer_file" = "$CHANGED_FILE" ] && continue
  [ ! -f "$peer_file" ] && continue
  # Check for identical code blocks (≥5 lines)
  awk '/^```/{p=!p;if(p){b=$0;c=1}else{b=b"\n"$0;if(c>=5)print b;b=""}}p&&c++' "$CHANGED_FILE" | \
  while IFS= read -r block; do
    if grep -Fq "$block" "$peer_file" 2>/dev/null; then
      echo "❌ DUPLICATE block in $peer_file"
    fi
  done
done
```
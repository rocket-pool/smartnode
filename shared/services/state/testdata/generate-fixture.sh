#!/bin/sh
# Captures a live NetworkState snapshot and updates network_state.json.gz.
# The existing fixture is archived to previous/network_state-slot-{slot}.json.gz
# before being replaced. If capture or validation fails, nothing is changed.
#
# Usage:
#   ./generate-fixture.sh -e <el-url> -b <bn-url> [-network mainnet|testnet|devnet] [-slot <slot>]
#
# Examples:
#   ./generate-fixture.sh -e http://localhost:8545 -b http://localhost:5052
#   ./generate-fixture.sh -e http://51.89.192.240:8888 -b http://51.89.192.240:5052 -slot 14800000

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
FIXTURE="$SCRIPT_DIR/network_state.json.gz"
TMPFILE="$SCRIPT_DIR/network_state.json.gz.tmp"
PREVIOUS_DIR="$SCRIPT_DIR/previous"

EL=""
BN=""
NETWORK="mainnet"
SLOT_ARG=""

usage() {
    echo "Usage: $0 -e <el-url> -b <bn-url> [-network mainnet|testnet|devnet] [-slot <slot>]" >&2
    echo "" >&2
    echo "Required:" >&2
    echo "  -e  Execution layer JSON-RPC URL (e.g. http://localhost:8545)" >&2
    echo "  -b  Beacon node REST API URL     (e.g. http://localhost:5052)" >&2
    echo "" >&2
    echo "Optional:" >&2
    echo "  -network  Network name: mainnet (default), testnet, or devnet" >&2
    echo "  -slot     Beacon slot to snapshot (default: head slot)" >&2
    echo "  -h        Show this help" >&2
    exit 1
}

# Parse flags
while [ $# -gt 0 ]; do
    case "$1" in
        -e) EL="$2"; shift 2 ;;
        -b) BN="$2"; shift 2 ;;
        -network) NETWORK="$2"; shift 2 ;;
        -slot) SLOT_ARG="-slot $2"; shift 2 ;;
        -h|--help) usage ;;
        *) echo "Unknown flag: $1" >&2; echo "" >&2; usage ;;
    esac
done

if [ -z "$EL" ] || [ -z "$BN" ]; then
    usage
fi

cleanup() {
    rm -f "$TMPFILE"
}
trap cleanup EXIT

echo "Fetching NetworkState from $NETWORK (EL: $EL, BN: $BN)..." >&2
echo "This may take a few minutes..." >&2

go run "$REPO_ROOT/shared/services/state/cli/" \
    -e "$EL" \
    -b "$BN" \
    -network "$NETWORK" \
    $SLOT_ARG \
    | gzip > "$TMPFILE"

echo "Validating fixture..." >&2
SLOT=$(zcat "$TMPFILE" | python3 -c "
import json, sys
d = json.load(sys.stdin)
assert 'network_details' in d, 'missing network_details'
assert 'beacon_slot_number' in d, 'missing beacon_slot_number'
validators = d.get('megapool_validator_global_index', [])
print(f'  beacon_slot:  {d[\"beacon_slot_number\"]}', file=sys.stderr)
print(f'  megapools:    {len(d.get(\"megapool_details\", {}))}', file=sys.stderr)
print(f'  validators:   {len(validators)}', file=sys.stderr)
print(f'  reduced_bond: {d[\"network_details\"][\"reduced_bond\"]}', file=sys.stderr)
print(d['beacon_slot_number'])
")

# Archive the existing fixture before replacing it
if [ -f "$FIXTURE" ]; then
    OLD_SLOT=$(zcat "$FIXTURE" | python3 -c "import json,sys; print(json.load(sys.stdin)['beacon_slot_number'])")
    mkdir -p "$PREVIOUS_DIR"
    ARCHIVE="$PREVIOUS_DIR/network_state-slot-$OLD_SLOT.json.gz"
    mv "$FIXTURE" "$ARCHIVE"
    echo "Archived previous fixture: $ARCHIVE" >&2
fi

mv "$TMPFILE" "$FIXTURE"
echo "Done. Fixture updated to slot $SLOT: $FIXTURE" >&2

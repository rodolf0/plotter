#!/bin/sh

__plot() {
  curl --data-binary @- "$@" http://localhost:7272/plot
}

if [[ "$*" =~ help ]]; then
  echo "usage: plot [--histo|--line] [--ws=<n>]" >&2
  exit 1
fi

PARAMS=()
GRAPH=()
WS=1
for param in "$@"; do
  case "$param" in
    --histo) GRAPH=(-H 'Graph: histChart') ;;
    --line) GRAPH=(-H 'Graph: lineChart') ;;
    --ws=*) WS="${param#--ws=}" ;;
    *) PARAMS=("${PARAMS[@]}" "$param") ;;
  esac
done

__plot -H "Workspace: $WS" "${GRAPH[@]}" "${PARAMS[@]}"

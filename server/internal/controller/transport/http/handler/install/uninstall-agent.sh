#!/bin/sh
set -eu

agent_path=/usr/local/bin/netstamp-agent

usage() {
	cat <<'EOF'
Usage:
  sh uninstall-agent.sh [--purge]

Runs:
  sudo netstamp-agent service uninstall [--purge]
EOF
}

die() {
	printf '%s\n' "$*" >&2
	exit 1
}

for arg in "$@"; do
	case "$arg" in
	--help)
		usage
		exit 0
		;;
	--purge)
		;;
	*)
		usage >&2
		die "unknown argument: $arg"
		;;
	esac
done

[ "$(id -u)" -eq 0 ] || die "this uninstaller must be run as root"

if [ ! -x "$agent_path" ]; then
	die "netstamp-agent is not installed at ${agent_path}; run netstamp-agent service uninstall from the installed binary"
fi

exec "$agent_path" service uninstall "$@"

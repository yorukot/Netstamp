#!/bin/sh
set -eu

install_path=/usr/local/bin/netstamp-agent
controller_url="__NETSTAMP_CONTROLLER_URL__"

usage() {
	cat <<'EOF'
Usage:
  sh agent.sh

Installs the netstamp-agent binary. Configure and start the systemd service with:
  sudo netstamp-agent service install --url CONTROLLER_URL --probe-id PROBE_ID --probe-secret PROBE_SECRET

Update the installed binary later with:
  sudo netstamp-agent update
EOF
}

die() {
	printf '%s\n' "$*" >&2
	exit 1
}

download() {
	url=$1
	dest=$2

	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url" -o "$dest"
		return
	fi

	if command -v wget >/dev/null 2>&1; then
		wget -qO "$dest" "$url"
		return
	fi

	die "curl or wget is required"
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--help)
		usage
		exit 0
		;;
	*)
		usage >&2
		die "unknown argument: $1"
		;;
	esac
done

[ "$(id -u)" -eq 0 ] || die "this installer must be run as root"
[ "$(uname -s)" = "Linux" ] || die "this installer supports Linux only"

case "$(uname -m)" in
x86_64 | amd64)
	binary_filename=netstamp-agent-linux-amd64
	;;
aarch64 | arm64)
	binary_filename=netstamp-agent-linux-arm64
	;;
*)
	die "this installer currently supports linux/amd64 and linux/arm64 only"
	;;
esac

binary_url="__NETSTAMP_AGENT_BINARY_BASE_URL__/${binary_filename}"

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

download "$binary_url" "${tmp_dir}/netstamp-agent"

install -d -m 0755 "$(dirname "$install_path")"
install -m 0755 "${tmp_dir}/netstamp-agent" "$install_path"

printf 'netstamp-agent installed at %s\n' "$install_path"
printf 'Configure the service with:\n'
printf '  sudo netstamp-agent service install --url %s --probe-id <probe-id> --probe-secret <probe-secret>\n' "$controller_url"
printf 'Update the installed binary later with:\n'
printf '  sudo netstamp-agent update\n'

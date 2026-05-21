#!/bin/sh
set -eu

service_name=netstamp-agent
system_user=netstamp
system_group=netstamp
home_dir=/var/lib/netstamp
install_path=/usr/local/bin/netstamp-agent
config_dir=/etc/netstamp
env_file="${config_dir}/probe.env"
service_file="/etc/systemd/system/${service_name}.service"
purge=0

usage() {
	cat <<'EOF'
Usage:
  sh uninstall-agent.sh [--purge]

Options:
  --purge  Remove /etc/netstamp/probe.env, /etc/netstamp, /var/lib/netstamp, and the netstamp system user/group.
  --help   Show this help
EOF
}

die() {
	printf '%s\n' "$*" >&2
	exit 1
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--purge)
		purge=1
		shift
		;;
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

[ "$(id -u)" -eq 0 ] || die "this uninstaller must be run as root"
[ "$(uname -s)" = "Linux" ] || die "this uninstaller supports Linux only"
command -v systemctl >/dev/null 2>&1 || die "systemctl is required"

if systemctl list-unit-files "$service_name.service" >/dev/null 2>&1 || [ -f "$service_file" ]; then
	systemctl disable --now "$service_name" >/dev/null 2>&1 || true
fi

rm -f "$service_file"
rm -f "$install_path"

systemctl daemon-reload
systemctl reset-failed "$service_name" >/dev/null 2>&1 || true

if [ "$purge" -eq 1 ]; then
	rm -f "$env_file"
	rmdir "$config_dir" 2>/dev/null || true
	rm -rf "$home_dir"

	if id -u "$system_user" >/dev/null 2>&1; then
		if command -v userdel >/dev/null 2>&1; then
			userdel "$system_user" 2>/dev/null || true
		elif command -v deluser >/dev/null 2>&1; then
			deluser "$system_user" 2>/dev/null || true
		fi
	fi

	if getent group "$system_group" >/dev/null 2>&1; then
		if command -v groupdel >/dev/null 2>&1; then
			groupdel "$system_group" 2>/dev/null || true
		elif command -v delgroup >/dev/null 2>&1; then
			delgroup "$system_group" 2>/dev/null || true
		fi
	fi
fi

printf 'Netstamp agent uninstalled.\n'
if [ "$purge" -ne 1 ] && [ -f "$env_file" ]; then
	printf 'Configuration preserved at %s. Re-run with --purge to remove it.\n' "$env_file"
fi

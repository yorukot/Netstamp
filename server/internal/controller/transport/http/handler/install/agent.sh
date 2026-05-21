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

api_version=v1
controller_url=
probe_id=
probe_secret=

usage() {
	cat <<'EOF'
Usage:
  sh agent.sh --controller-url URL --probe-id UUID --probe-secret SECRET [--api-version VERSION]

Options:
  --controller-url URL   Netstamp controller origin, for example https://netstamp.example.com
  --probe-id UUID        Probe ID from the Netstamp project
  --probe-secret SECRET  Plaintext probe secret from creation or rotation
  --api-version VERSION  API version used by the controller install endpoint (default: v1)
  --help                 Show this help
EOF
}

die() {
	printf '%s\n' "$*" >&2
	exit 1
}

require_arg() {
	name=$1
	value=${2-}
	if [ -z "$value" ]; then
		die "${name} requires a value"
	fi
}

reject_newline() {
	name=$1
	value=$2
	case "$value" in
	*'
'*) die "${name} must not contain a newline" ;;
	esac
}

systemd_env_escape() {
	printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g; s/\$/\\$/g; s/`/\\`/g'
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

nologin_shell() {
	if [ -x /usr/sbin/nologin ]; then
		printf '%s\n' /usr/sbin/nologin
		return
	fi
	if [ -x /sbin/nologin ]; then
		printf '%s\n' /sbin/nologin
		return
	fi
	printf '%s\n' /bin/false
}

ensure_group() {
	if getent group "$system_group" >/dev/null 2>&1; then
		return
	fi

	if command -v groupadd >/dev/null 2>&1; then
		groupadd --system "$system_group"
		return
	fi

	if command -v addgroup >/dev/null 2>&1; then
		addgroup --system "$system_group" 2>/dev/null || addgroup -S "$system_group"
		return
	fi

	die "groupadd or addgroup is required"
}

ensure_user() {
	if id -u "$system_user" >/dev/null 2>&1; then
		return
	fi

	shell=$(nologin_shell)
	if command -v useradd >/dev/null 2>&1; then
		useradd --system --gid "$system_group" --home-dir "$home_dir" --create-home --shell "$shell" "$system_user"
		return
	fi

	if command -v adduser >/dev/null 2>&1; then
		adduser --system --ingroup "$system_group" --home "$home_dir" --shell "$shell" --disabled-login --gecos "" "$system_user" 2>/dev/null \
			|| adduser -S -D -h "$home_dir" -s "$shell" -G "$system_group" "$system_user"
		return
	fi

	die "useradd or adduser is required"
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--controller-url)
		require_arg "$1" "${2-}"
		controller_url=$2
		shift 2
		;;
	--probe-id)
		require_arg "$1" "${2-}"
		probe_id=$2
		shift 2
		;;
	--probe-secret)
		require_arg "$1" "${2-}"
		probe_secret=$2
		shift 2
		;;
	--api-version)
		require_arg "$1" "${2-}"
		api_version=$2
		shift 2
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

[ "$(id -u)" -eq 0 ] || die "this installer must be run as root"
[ "$(uname -s)" = "Linux" ] || die "this installer supports Linux only"
command -v systemctl >/dev/null 2>&1 || die "systemctl is required"
[ -d /run/systemd/system ] || die "systemd does not appear to be running"
[ -n "$controller_url" ] || die "--controller-url is required"
[ -n "$probe_id" ] || die "--probe-id is required"
[ -n "$probe_secret" ] || die "--probe-secret is required"
[ -n "$api_version" ] || die "--api-version is required"

reject_newline "--controller-url" "$controller_url"
reject_newline "--probe-id" "$probe_id"
reject_newline "--probe-secret" "$probe_secret"
reject_newline "--api-version" "$api_version"

controller_url=${controller_url%/}
binary_url="${controller_url}/api/${api_version}/install/netstamp-agent-linux-amd64"

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

download "$binary_url" "${tmp_dir}/netstamp-agent"

ensure_group
ensure_user

install -d -m 0755 "$(dirname "$install_path")"
install -m 0755 "${tmp_dir}/netstamp-agent" "$install_path"

install -d -m 0755 "$config_dir"
umask 077
{
	printf '# Managed by the Netstamp agent installer.\n'
	printf 'NETSTAMP_PROBE_CONTROLLER_URL="%s"\n' "$(systemd_env_escape "$controller_url")"
	printf 'NETSTAMP_PROBE_ID="%s"\n' "$(systemd_env_escape "$probe_id")"
	printf 'NETSTAMP_PROBE_SECRET="%s"\n' "$(systemd_env_escape "$probe_secret")"
	printf 'NETSTAMP_PROBE_LOG_LEVEL="info"\n'
} >"$env_file"
chmod 0600 "$env_file"
chown root:root "$env_file"

cat >"$service_file" <<EOF
[Unit]
Description=Netstamp Probe Agent
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=${system_user}
Group=${system_group}
EnvironmentFile=${env_file}
ExecStart=${install_path}
Restart=always
RestartSec=5s
AmbientCapabilities=CAP_NET_RAW
CapabilityBoundingSet=CAP_NET_RAW
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=full

[Install]
WantedBy=multi-user.target
EOF

chmod 0644 "$service_file"

systemctl daemon-reload
systemctl enable --now "$service_name"

printf 'Netstamp agent installed and started.\n'
printf 'Logs: journalctl -u %s -f\n' "$service_name"

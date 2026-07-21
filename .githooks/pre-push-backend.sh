#!/bin/sh

set -eu

refs_file=${1:?"Usage: pre-push-backend.sh <pre-push-refs-file>"}
zero_sha="0000000000000000000000000000000000000000"
generated_openapi_path="server/internal/controller/transport/http/openapi/openapi.json"
has_backend_changes=0

has_backend_paths() {
	if [ -n "$(printf '%s\n' "$1" | awk -v generated_openapi_path="$generated_openapi_path" '
		$0 == generated_openapi_path {
			next
		}
		/^server\// || $0 == "golangci.yaml" {
			print
			exit
		}
	')" ]; then
		return 0
	fi

	return 1
}

ref_has_backend_changes() {
	local_sha=$1
	remote_sha=$2

	if [ "$local_sha" = "$zero_sha" ]; then
		return 1
	fi

	if [ "$remote_sha" != "$zero_sha" ]; then
		if changed_files=$(git diff --name-only "$remote_sha" "$local_sha" -- server golangci.yaml 2>/dev/null); then
			has_backend_paths "$changed_files"
			return $?
		fi

		echo "Could not compare pushed commits; running backend checks to be safe." >&2
		return 0
	fi

	if commits=$(git rev-list "$local_sha" --not --remotes 2>/dev/null); then
		if [ -z "$commits" ]; then
			return 1
		fi

		for commit in $commits; do
			changed_files=$(git diff-tree --no-commit-id --name-only -r -m --root "$commit" -- server golangci.yaml)
			if has_backend_paths "$changed_files"; then
				return 0
			fi
		done

		return 1
	fi

	echo "Could not inspect pushed commits; running backend checks to be safe." >&2
	return 0
}

while read -r _local_ref local_sha _remote_ref remote_sha; do
	if ref_has_backend_changes "$local_sha" "$remote_sha"; then
		has_backend_changes=1
		break
	fi
done <"$refs_file"

if [ "$has_backend_changes" -eq 0 ]; then
	echo "No backend changes detected in pushed commits; skipping Go checks."
	exit 0
fi

if ! command -v golangci-lint >/dev/null 2>&1; then
	echo "golangci-lint is required for backend pre-push checks." >&2
	echo "Install golangci-lint v2.12.2, then retry the push." >&2
	exit 1
fi

required_golangci_lint_version="2.12.2"
actual_golangci_lint_version=$(golangci-lint version --short)
if [ "$actual_golangci_lint_version" != "$required_golangci_lint_version" ]; then
	echo "Backend checks require golangci-lint $required_golangci_lint_version; found $actual_golangci_lint_version." >&2
	echo "Install the required version, then retry the push." >&2
	exit 1
fi

echo "Backend changes detected; checking Go formatting..."
if ! (cd server && golangci-lint fmt --config ../golangci.yaml --diff); then
	echo "Backend formatting check failed. Run: just backend-fmt" >&2
	exit 1
fi

echo "Checking backend Go lint..."
if ! (cd server && golangci-lint run --config ../golangci.yaml ./...); then
	echo "Backend lint failed. Run: just backend-lint" >&2
	exit 1
fi

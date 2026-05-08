#!/bin/sh

set -eu

refs_file=${1:?"Usage: pre-push-frontend.sh <pre-push-refs-file>"}
zero_sha="0000000000000000000000000000000000000000"
has_frontend_changes=0

has_frontend_paths() {
	if [ -n "$(printf '%s\n' "$1" | sed -n '/^web\//p; /^packages\/ui\//p; /^packages\/brand\//p; /^package\.json$/p; /^pnpm-lock\.yaml$/p; /^pnpm-workspace\.yaml$/p; /^\.prettierrc$/p; /^\.prettierignore$/p' | head -n 1)" ]; then
		return 0
	fi

	return 1
}

ref_has_frontend_changes() {
	local_sha=$1
	remote_sha=$2

	if [ "$local_sha" = "$zero_sha" ]; then
		return 1
	fi

	if [ "$remote_sha" != "$zero_sha" ]; then
		if changed_files=$(git diff --name-only "$remote_sha" "$local_sha" -- web packages/ui packages/brand package.json pnpm-lock.yaml pnpm-workspace.yaml .prettierrc .prettierignore 2>/dev/null); then
			has_frontend_paths "$changed_files"
			return $?
		fi

		echo "Could not compare pushed commits; running frontend checks to be safe." >&2
		return 0
	fi

	if commits=$(git rev-list "$local_sha" --not --remotes 2>/dev/null); then
		if [ -z "$commits" ]; then
			return 1
		fi

		for commit in $commits; do
			changed_files=$(git diff-tree --no-commit-id --name-only -r -m --root "$commit" -- web packages/ui packages/brand package.json pnpm-lock.yaml pnpm-workspace.yaml .prettierrc .prettierignore)
			if has_frontend_paths "$changed_files"; then
				return 0
			fi
		done

		return 1
	fi

	echo "Could not inspect pushed commits; running frontend checks to be safe." >&2
	return 0
}

while read -r _local_ref local_sha _remote_ref remote_sha; do
	if ref_has_frontend_changes "$local_sha" "$remote_sha"; then
		has_frontend_changes=1
		break
	fi
done <"$refs_file"

if [ "$has_frontend_changes" -eq 0 ]; then
	echo "No frontend changes detected in pushed commits; skipping frontend checks."
	exit 0
fi

if ! command -v pnpm >/dev/null 2>&1; then
	echo "pnpm is required for frontend pre-push checks." >&2
	exit 1
fi

echo "Frontend changes detected; checking formatting..."
if ! pnpm prettier --check web packages/ui packages/brand --ignore-unknown; then
	echo "Frontend formatting check failed. Run: pnpm format" >&2
	exit 1
fi

echo "Checking web lint..."
if ! pnpm --filter @netstamp/web lint; then
	echo "Web lint failed. Run: pnpm --filter @netstamp/web lint" >&2
	exit 1
fi

echo "Checking web typecheck..."
if ! pnpm --filter @netstamp/web typecheck; then
	echo "Web typecheck failed. Run: pnpm --filter @netstamp/web typecheck" >&2
	exit 1
fi

echo "Checking shared UI typecheck..."
if ! pnpm --filter @netstamp/ui typecheck; then
	echo "Shared UI typecheck failed. Run: pnpm --filter @netstamp/ui typecheck" >&2
	exit 1
fi

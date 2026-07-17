#!/bin/sh

set -eu

valid_pattern='^(feat|fix|ui|refactor|docs|test|chore|release)/[a-z0-9]+(-[a-z0-9]+)*$'

validate_branch_name() {
	branch_name=$1

	if [ "$branch_name" = "main" ]; then
		return 0
	fi

	if printf '%s\n' "$branch_name" | LC_ALL=C grep -Eq "$valid_pattern"; then
		return 0
	fi

	printf 'Invalid branch name: %s\n' "$branch_name" >&2
	printf '%s\n' 'Expected <type>/<short-kebab-case-description>.' >&2
	printf '%s\n' 'Allowed types: feat, fix, ui, refactor, docs, test, chore, release.' >&2
	return 1
}

if [ "$#" -eq 0 ]; then
	current_branch=$(git symbolic-ref --quiet --short HEAD 2>/dev/null || true)
	if [ -z "$current_branch" ]; then
		printf '%s\n' 'Detached HEAD; no branch name to validate.'
		exit 0
	fi

	set -- "$current_branch"
fi

for branch_name in "$@"; do
	validate_branch_name "$branch_name"
done

#!/usr/bin/env bash
# merged_branches.sh — list local branches and whether each is fully merged into master.
#
# Usage: ./scripts/merged_branches.sh [-clean]
#
# Branches are sorted ascending by last-commit time. The current branch is
# marked with "*". "merged" is yes/no/- (- for master itself).
#
# -clean   delete every local branch fully merged into master, excluding master
#          and the current branch. Uses `git branch -d` (refuses unmerged).

set -euo pipefail

CLEAN=0

for arg in "$@"; do
    case "$arg" in
        -clean) CLEAN=1 ;;
        -h|--help)
            sed -n '2,/^set -euo/p' "$0" | sed '$d' | sed 's|^# \?||'
            exit 0
            ;;
        *)
            echo "merged_branches.sh: unknown argument: $arg" >&2
            exit 2
            ;;
    esac
done

if ! git rev-parse --git-dir >/dev/null 2>&1; then
    echo "merged_branches.sh: not a git repository" >&2
    exit 1
fi

if git show-ref --verify --quiet refs/heads/master; then
    MAIN=master
elif git show-ref --verify --quiet refs/heads/main; then
    MAIN=main
else
    echo "merged_branches.sh: no master or main branch found" >&2
    exit 1
fi

CURRENT=$(git symbolic-ref --short HEAD 2>/dev/null || echo "")

if [ "$CLEAN" = 1 ]; then
    echo "==> -clean: deleting branches merged into $MAIN"
    while IFS= read -r b; do
        [ -z "$b" ] && continue
        [ "$b" = "$MAIN" ] && continue
        if [ "$b" = "$CURRENT" ]; then
            echo "  skip $b (current branch)"
            continue
        fi
        if git branch -d "$b"; then
            echo "  deleted $b"
        else
            echo "  warning: failed to delete $b" >&2
        fi
    done < <(git branch --merged "$MAIN" --format='%(refname:short)')
    echo
fi

declare -A MERGED
while IFS= read -r b; do
    [ -n "$b" ] && MERGED["$b"]=1
done < <(git branch --merged "$MAIN" --format='%(refname:short)')

printf "%-25s %-50s %-8s %s\n" "LAST_COMMIT" "BRANCH" "MERGED" ""
git for-each-ref refs/heads/ \
    --sort=committerdate \
    --format='%(committerdate:iso-strict)|%(refname:short)' |
while IFS='|' read -r date branch; do
    if [ "$branch" = "$MAIN" ]; then
        merged="-"
    elif [ -n "${MERGED[$branch]:-}" ]; then
        merged="yes"
    else
        merged="no"
    fi
    mark=""
    [ "$branch" = "$CURRENT" ] && mark="*"
    printf "%-25s %-50s %-8s %s\n" "$date" "$branch" "$merged" "$mark"
done

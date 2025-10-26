#!/bin/bash
set -euo pipefail

commit_message_check() {
    # Ensure we have main available
    if ! git rev-parse --verify origin/main >/dev/null 2>&1; then
        echo "Fetching main branch..."
        git fetch origin main:refs/remotes/origin/main
    fi

    # Get commits in the current HEAD not in main
    git log HEAD --pretty=format:"%H" --not origin/main > shafile.txt

    failed=0

    while read -r sha; do
        git log --format=%B -n 1 "$sha" > msgfile.txt

        if ! egrep '^bench [0-9]+$' msgfile.txt > /dev/null; then
            echo "Your commit message must contain the bench number"
            echo "it was :"
            cat msgfile.txt
            echo "--- end ---"
            failed=1
        fi
    done < shafile.txt

    rm -f shafile.txt msgfile.txt >/dev/null 2>&1

    exit $failed
}

commit_message_check


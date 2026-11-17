#!/bin/sh
# Ctrl-C check for masked cred run. Run setup.sh first, run this in a
# terminal, and press Ctrl-C once when "press Ctrl-C" appears.
KIT="${XDG_CACHE_HOME:-$HOME/.cache}/cred-manual-check"
export DEMO_TOKEN=cred://cred-manual-check/DEMO_TOKEN

echo "== 1/2: plain command (sleep 30). Press Ctrl-C; it should stop at once."
"$KIT/cred-dev" run -- sh -c 'echo "token=$DEMO_TOKEN"; echo "press Ctrl-C"; exec sleep 30'
echo "-> exit code $? (expect 130)"
echo

echo "== 2/2: command with a SIGINT handler. Press Ctrl-C; it should print got SIGINT and stop."
"$KIT/cred-dev" run -- sh -c 'trap "echo got SIGINT; exit 7" INT; echo "token=$DEMO_TOKEN"; echo "press Ctrl-C"; while :; do sleep 1; done'
echo "-> exit code $? (expect 7)"

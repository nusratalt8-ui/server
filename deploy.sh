#!/bin/bash

HOST="199.231.191.240"
USER="root"
PASS="7OfGTj8NwZx5"
REMOTE_DIR="/root/agentpanel"
PRODIGNORE=".prodignore"

EXCLUDE_ARGS=""
while IFS= read -r line; do
    [[ -z "$line" || "$line" == \#* ]] && continue
    EXCLUDE_ARGS="$EXCLUDE_ARGS --exclude=$line"
done < "$PRODIGNORE"
EXCLUDE_ARGS="$EXCLUDE_ARGS --exclude=.prodignore --exclude=deploy.sh --exclude=ssh.sh"

sshpass -p "$PASS" rsync -avz $EXCLUDE_ARGS \
    -e "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
    ./ "$USER@$HOST:$REMOTE_DIR/"

sshpass -p "$PASS" rsync -avz \
    -e "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
    ./data/schema.sql "$USER@$HOST:$REMOTE_DIR/data/schema.sql"

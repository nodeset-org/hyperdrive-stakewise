#!/usr/bin/env bash 



echo "Cleaning up previous configuration..."

# if a configuration exists, shut it down first
if [ -f "$DATA_DIR/nodeset.env" ]; then
    "$SCRIPT_DIR/nodeset.sh" "-d" "$DATA_DIR" "shutdown"
fi

# clear old data (if any)
echo "Deleting previous configuration at $DATA_DIR"
rm -rd "$DATA_DIR"
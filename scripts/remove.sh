#!/usr/bin/env bash 

confirm()
{
    echo "Are you sure you want to delete your previous configuration completely? (y/n)"
    read confirm
}

# if data directory exists and isn't empty, confirm deletion
if [ -d "$DATA_DIR" ] && [ -n "$(ls -A "$DATA_DIR")" ]; then
    confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "n" ]; then
        confirm
    elif [ "$confirm" = "n" ]; then
        echo "Cancelled"
        exit 1
    fi
fi

echo "Cleaning up previous configuration..."

# if a configuration exists, shut it down first
if [ -f "$DATA_DIR/nodeset.env" ]; then
    "$SCRIPT_DIR/nodeset.sh" "-d" "$DATA_DIR" "shutdown"
fi

# clear old data (if any)
echo "Deleting previous configuration at $DATA_DIR"
rm -rd "$DATA_DIR"
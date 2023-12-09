#!/usr/bin/env bash 

confirm()
{
    echo "Are you sure you want to delete your previous configuration completely? (y/n)"
    read confirm
}

if [ -d "$DATA_DIR" ] && [ -n "$(ls -A "$DATA_DIR")" ]; then
    confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "n" ]; then
        confirm
    elif [ "$confirm" = "n" ]; then
        echo "Cancelled"
        exit
    fi
fi

echo "Cleaning up previous configuration..."

"$SCRIPT_DIR/nodeset.sh" "-d" "$DATA_DIR" "shutdown"

# clear old data (if any)
sudo rm -rd $DATA_DIR
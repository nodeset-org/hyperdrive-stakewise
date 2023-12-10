#!/usr/bin/env bash 

if [ "$NETWORK" = "mainnet" ]; then

    # todo: also check if there are any active validators before giving this warning
    # i.e. docker compose up geth "check validators request"
    echo "DANGER: You are attempting to reset your configuration for a mainnet vault!"
    echo "This will require you to resync the chain completely before you can begin validating again, which may take several days."
    echo "Remember, if you're offline for too long, you may be kicked out of NodeSet!"
    echo
    echo "Are you sure you want to continue? You must type 'I UNDERSTAND' to continue."
    read answer

    if [ "$answer" != "I UNDERSTAND" ]; then 
        echo Cancelled
        exit
    fi

    echo "THIS IS YOUR FINAL WARNING! Are you absolutely sure that you want to delete all of your data for this mainnet configuration?"
    echo
    echo "You must type 'DELETE EVERYTHING' to continue."
    read answer2

    if [ "$answer2" != "DELETE EVERYTHING" ]; then 
        echo Cancelled
        exit
    fi
else
    read_input()
    {
        echo "Are you sure you want to delete your previous configuration completely? (y/n)"
        read confirm
    }

    confirm()
    {
        if [ "$confirm" != "y" ] && [ "$confirm" != "n" ]; then
            read_input
            confirm
        elif [ "$confirm" = "n" ]; then
            echo "Cancelled"
            exit 2
        fi
    }

    # if data directory already exists and isn't empty, confirm deletion
    if [ -d "$DATA_DIR" ] && [ -n "$(ls -A "$DATA_DIR")" ]; then
        confirm
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
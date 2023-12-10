#!/usr/bin/env bash 

## -- Start of Script -- ##

# check if bash
if [ "$BASH_VERSION" = '' ]; then
    printf "Please execute this with a bash-compatible shell.\nExample: sudo bash init-node.sh holesky"
    exit
fi

# ensure root access
if [ "$(id -u)" -ne 0 ]; then
  echo "Please run as root (or with sudo)."
  exit
fi

export SCRIPT_DIR=$( cd -- "$( dirname -- "$( realpath ${BASH_SOURCE[0]} )" )" &> /dev/null && pwd )
export APP_DIR=$( cd -- "$SCRIPT_DIR/.." &> /dev/null && pwd )
export DATA_DIR=""
usagemsg="Usage: nodeset [--help|-h] [--data-dir|-d=DATA_DIRECTORY] [COMMAND] \nCommands:\nlogs\t\tShow node logs\nshutdown\tShuts down the node\nremove\t\tCompletely deletes the existing installation\nstart\t\tStarts the node\n"
reset=false
shutdown=false
if [ $SUDO_USER ]; then 
    callinguser=$SUDO_USER; 
else 
    callinguser=`whoami`
fi

# check if installation
if [ "$1" = "install" ]; then
    "sudo $SCRIPT_DIR/install-node.sh"
fi

while getopts "hd:-:" option; do
    case $option in
        -)
            case "${OPTARG}" in
                help)
                    printf "$usagemsg\n"
                    exit 0
                    ;;
                data-directory=*)
                    export DATA_DIR=${OPTARG#*=}
                    ;;
                data-directory)
                    export DATA_DIR="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                \?)
                    printf "$usagemsg\n"
                    exit 1
                    ;;
                :)
                    printf "ERROR: Option -$option requires an argument\n\n"
                    printf "$usagemsg\n"
                    exit 1
                    ;;
                *) 
                    echo "Unknown arg -$option"
                    exit 1
                    ;;
            esac
            ;;
        h)
            printf "$usagemsg\n"
            exit 0
            ;;
        d)
            export DATA_DIR=${OPTARG}
            ;;
        d=*)
            export DATA_DIR=${OPTARG#*=}
            ;;
        \?)
            printf "$usagemsg\n" 
            exit 1
            ;;
        :)
            printf "ERROR: Option -$option requires an argument\n\n"
            printf "$usagemsg\n"
            exit 1
            ;;
        *) 
            echo "Unknown arg"
            exit 1
            ;;
    esac
done
shift $(( OPTIND - 1 ))

if [ "$1" != remove ]; then
    if [ ! -d "$DATA_DIR" ] || [ ! -f "$DATA_DIR/nodeset.env" ]; then
        echo "No installation found. Please run the installer using \"sudo bash install-node.sh\" or check to make sure the correct data directory was provided."
        exit
    fi
fi

# set env based on installation config
# only do this if not removing or displaying help down the node
if [ "$1" != "remove" ] && [ "$1" != "help" ]; then
    if [ -f "$DATA_DIR/nodeset.env" ]; then
        set -a 
        source "$DATA_DIR/nodeset.env"
        set +a
    else
        echo "FATAL ERROR: Cannot find nodeset.env configuration file"
        echo "Are you sure this data directory is correct? If so, you must recover your configuration manually."
        echo "Given data directory: $DATA_DIR/nodeset.env"
        exit 2
    fi
fi

remove()
{
    if [ "$NETWORK" = "mainnet" ]; then

        # todo: check if there are any active validators before giving this warning
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
    
    "$SCRIPT_DIR/remove.sh"
}


# check command name makes sense
case "$1" in
    help)
        printf "$usagemsg\n"
        exit
        ;;
    remove)
        remove
        ;;
    shutdown)
        "$SCRIPT_DIR/shutdown.sh"
        exit
        ;;
    start)
        "$SCRIPT_DIR/start.sh"
        ;;
    logs)
        docker compose -f "$SCRIPT_DIR/compose.yaml" logs -f
        ;;
    *)
        printf "You must provide a command!\n\n"
        printf "$usagemsg\n"
        exit
        ;;
esac



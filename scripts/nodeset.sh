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

export SCRIPT_DIR=$( dirname -- "$( readlink -f -- "${BASH_SOURCE[0]}"; )"; )
set -a
source "$SCRIPT_DIR/scripts.env"
set +a
export DATA_DIR=""
usagemsg="Usage: nodeset [--help|-h] [--data-dir|-d=DATA_DIRECTORY] [COMMAND] \nCommands:\nlogs\t\tShow node logs\nshutdown\tShuts down the node\nremove\t\tCompletely deletes the existing installation\nstart\t\tStarts the node\n"
reset=false
shutdown=false
if [ $SUDO_USER ]; then 
    callinguser=$SUDO_USER; 
else 
    callinguser=`whoami`
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

if [ "$1" != "remove" ] && [ "$1" != "help" ]; then
    if [ ! -d "$DATA_DIR" ] || [ ! -f "$DATA_DIR/nodeset.env" ]; then
        echo "No installation found. Please run the installer using \"sudo bash install-node.sh\" or check to make sure the correct data directory was provided."
        exit
    fi
fi

# set env based on installation config
if [ "$1" != "help" ]; then
    if [ -f "$DATA_DIR/nodeset.env" ]; then
        set -a 
        source "$DATA_DIR/nodeset.env"
        set +a
    elif [ "$1" != "remove" ]; then # don't show an error if we're removing anyway
        echo "FATAL ERROR: Cannot find nodeset.env configuration file"
        echo "Are you sure this data directory is correct? If so, you must recover your configuration manually."
        echo "Given data directory: $DATA_DIR/nodeset.env"
        exit 2
    fi
fi

# check command name makes sense
case "$1" in
    help)
        printf "$usagemsg\n"
        exit
        ;;
    remove)
        "$SCRIPT_DIR/remove.sh"
        ;;
    shutdown)
        echo "Shutting down containers..."
        docker compose -f "$DATA_DIR/compose.yaml" down
        exit
        ;;
    start)
        "$SCRIPT_DIR/start.sh"
        ;;
    logs)
        docker compose -f "$DATA_DIR/compose.yaml" logs -f
        ;;
    *)
        printf "You must provide a command!\n\n"
        printf "$usagemsg\n"
        exit
        ;;
esac



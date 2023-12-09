#!/usr/bin/env bash 

## -- Start of Script -- ##

# check if bash
if [ "$BASH_VERSION" = '' ]; then
    printf "Please execute this with a bash-compatible shell.\nExample: sudo bash init-node.sh holesky"
    exit
fi

# ensure root access
if [ "$(id -u)" -ne 0 ]; then
  echo "Please run as root (or with sudo)"
  exit
fi

SCRIPT_DIR=$( cd -- "$( dirname -- "$( realpath ${BASH_SOURCE[0]} )" )" &> /dev/null && pwd )
usagemsg="Usage: nodeset [COMMAND] \nCommands:\nremove -- Completely deletes the existing installation\nrun [--data-dir|-d=DATA_DIRECTORY]"
reset=false
shutdown=false
data_dir=""
if [ $SUDO_USER ]; then 
    callinguser=$SUDO_USER; 
else 
    callinguser=`whoami`
fi

# check if installation
if [ "$1" = "install" ]; then
    "sudo $SCRIPT_DIR/node-install.sh"
fi

while getopts "h-:" option; do
    case $option in
        -)
            case "${OPTARG}" in
                help)
                    printf "$usagemsg\n"
                    exit 0
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


if [ -d "$data_dir" ]; then
    if [ -f "$data_dir/*.env")" ]; then  
        
    fi
else
    echo "No existing installation found. Exiting."
    exit
fi
echo "No installation found. Please run the installer using `bash node-install.sh` or check to make sure the correct data directory was provided."
# if yes, get vault config automatically


# check command name makes sense
case "$1" in
    remove)
        echo "remove command found"
        ;;
    shutdown)
        echo "shutdown command found"
        ;;
    run)
        echo "run command found"
        ;;
    *)
        echo "ERROR: incorrect command"
        printf "$usagemsg\n"
        exit
        ;;
esac

# set env based on vault installation
set -a 
source $data_dir/$vault.env
set +a

if [ "$shutdown" = true ]; then
    "$SCRIPT_DIR/shutdown.sh"
    exit
fi

if [ "$reset" = true ]; then
     if [ "$1" = "mainnet" ]; then

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
    fi
    "$SCRIPT_DIR/remove.sh"
fi

#!/usr/bin/env bash 

## -- Start of Script -- ##

# check if bash
if [ "$BASH_VERSION" = '' ]; then
    printf "Please execute this with a bash-compatible shell.\nExample: sudo bash init-node.sh holesky"
    exit
fi

# ensure root access
if [ "$(id -u)" -ne 0 ]; then
  echo "Please run this script as root (or with sudo)"
  exit
fi

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
usagemsg="Usage: init-node.sh [--reset|-r] [--mnemonic|-m=mnemonic] VAULT\nSupported vaults: holesky, gravita\nExample: sudo sh init-node.sh -m \"correct horse battery staple...\" holesky"
reset=false
shutdown=false
data_dir=/home/${ whoami }/node-data

# check if installation
if [ "$1" = "install" ]
    sudo bash install.sh
fi

while getopts "rhsd:m:-:" option; do
    case $option in
        -)
            case "${OPTARG}" in
                reset)
                    reset=true
                    ;;
                mnemonic=*)
                    mnemonic=${OPTARG#*=}
                    ;;
                mnemonic)
                    mnemonic="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                data-directory=*)
                    data_dir=${OPTARG#*=}
                    ;;
                data-directory)
                    data_dir="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                shutdown)
                    shutdown=true
                    ;;
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
        d)
            data_dir=${OPTARG}
            ;;
        r)
            reset=true
            ;;
        s)
            shutdown=true
            ;;
        h)
            printf "$usagemsg\n"
            exit 0
            ;;
        m)
            mnemonic=${OPTARG}
            ;;
        m=*)
            mnemonic=${OPTARG#*=}
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


# if no, install

# if yes, get vault config automatically


# check vault name makes sense
if [ "$1" != "holesky" ] && [ "$1" != "gravita" ]; then
    printf "Error: you must provide a valid vault name\n\n"
    printf "$usagemsg\n"
    exit
fi

# set env based on vault installation
set -a 
source $data_dir/${1}.env
set +a

if [ "$shutdown" = true ]; then
    remove_containers
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
    sudo bash reset.sh
fi

#!/usr/bin/env bash 

if [ $ECNAME != "external" ]; then
    composeFile=(-f "$DATA_DIR/compose.yaml" -f "$DATA_DIR/compose.internal.yaml")
else
    composeFile=(-f "$DATA_DIR/compose.yaml")
fi

if [ "$1" = "" ]; then
    docker compose ${composeFile[@]} logs -f
    exit
fi

if [ "$1" != "$ECNAME" ] && [ "$1" != "$CCNAME" ] && [ "$1" != "stakewise" ] && [ "$1" != "ethdo" ]; then
    echo "$1 is not a valid container name."
    if [ $$ECNAME != "external" ]; then
        echo "Available options: $CCNAME, $ECNAME, stakewise, ethdo"
    else
        echo "Available options: stakewise, ethdo"
    fi
fi

docker compose ${composeFile[@]} logs "$1" -f
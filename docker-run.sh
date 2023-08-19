#!/bin/bash
# we execpt to have a
# - source network, 
# - a dst network
# - a dest addr in the ip:port format
# - a name for the container, in that order

if [ $# -ne 4 ]; then 
    echo "Expected parameters: <docker external net> <docker internal net> <docker ip> <name>"
else
    source_net=$1
    dest_net=$2
    dest_addr=$3
    name=$4
    
    go build
    docker build . -t goxii
    docker run --network $source_net \
        -itd \
        --name goxii-descendent goxii:latest \
        ./Goxii --descendent
    docker network connect internal goxii-descendent
    
    docker run --network host \
        -p 8080:8080 \
        -p 9000:9000 \
        -itd \
        --name $name goxii:latest \
        ./Goxii
    docker network connect $source_net $name 
fi
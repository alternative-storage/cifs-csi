#!/bin/sh

# usage:
#  net rpc share add $SHARE_NAME=$PATH_TO_SHARE $COMMENT

config_file=$1
share_name=$2
path_name=$3
share_path=$path_name/$share_name
comment=$4
maxconnections=$5

if [ ! -d $share_path ]; then
    mkdir -p $share_path
    chmod 755 $share_path
else
    echo "$path_name already exists"
fi

net usershare add $share_name $share_path $comment -M $maxconnections

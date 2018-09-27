#!/bin/sh

# net rpc share delete SHARE_NAME /PATH/TO/SHARE

config_file=$1
share_name=$2

if [ "$share_name" = "" ]
then
    exit 1
fi

if ! net usershare info $share_name > /dev/null
then
    exit 1
fi

net usershare delete $share_name

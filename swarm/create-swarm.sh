#!/bin/bash

./aws-cluster.sh -a ami-4bf3d731 \
    -v vpc-3e859d5a \
    -r us-east-1  \
    -z a \
    -i m4.xlarge \
    -s subnet-fd3db7d7 \
    -sg ems-swarm \
    -su centos \
    -m 1 \
    -w 1 \
    -p aws \
    -rp 5000
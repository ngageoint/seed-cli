#!/bin/bash

# Capture command line arguments
INPUT=$1
OUTPUT=$2

echo 'Printing working directory'
pwd

DIR=$(dirname $INPUT)
echo 'lsing dir of ' ${DIR}
ls -l $(dirname ${DIR})

# ls -l /home/docker/

echo ''
echo '----------------------------------------------------'
echo 'Calling job with arguments ' ${INPUT} ${OUTPUT}
SCRIPT=my_alg.py

python ${SCRIPT} ${INPUT} ${OUTPUT}
rc=$?
echo 'Done calling job - wrapper finished'
echo '----------------------------------------------------'
echo ''

echo 'lsing dir of ' ${OUTPUT}
ls -l ${OUTPUT}
exit ${rc}

#!/bin/bash

# Capture command line arguments
INPUT=$1
OUTPUT=$2

echo ''
echo '----------------------------------------------------'
echo 'Input file is  ' ${INPUT} 
echo 'output file is ' ${OUTPUT}
echo '----------------------------------------------------'
echo 'this is stdout'
sleep 2
>&2 echo 'this is stderr'
sleep 2
echo '----------------------------------------------------'
echo 'this is stdout again'
sleep 2
>&2 echo 'this is stderr again'
sleep 2
echo ''
touch ${OUTPUT}/my_output.txt
exit 0

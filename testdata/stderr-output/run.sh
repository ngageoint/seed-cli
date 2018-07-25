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
>&2 echo 'this is stderr'
echo '----------------------------------------------------'
echo ''
touch ${OUTPUT}/my_output.txt
exit 0

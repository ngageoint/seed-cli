#!/bin/bash

# Creates a docker-machine connection to an existing machine.
# Arguments: IP Address SSH User
usage() 
{
    echo 'Usage: aws-connect-machine [OPTIONS]'
    echo 'Options:'
    echo '-ip <The private IP of the AWS instance>'
    echo '-su <Default SSH user for machine>'
    echo '-n  <Name of the machine>'
    exit
}

if [ "$#" -ne 6 ]
then 
    usage
fi 
while [ "$1" != "" ]; do
case $1 in
    -ip )   shift
            IP=$1
            ;;
    -su )   shift
            SU=$1
            ;;
    -n )    shift
            NAME=$1
            ;;
esac
shift
done

echo "Connecting to "${IP}" as "${SU}

docker-machine create --driver generic --generic-ip-address=$1 --generic-ssh-user=${SU} ${NAME}
echo "${NAME} created"
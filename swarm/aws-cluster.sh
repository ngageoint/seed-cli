#!/bin/bash

usage() 
{
    echo 'Usage: aws-cluster [OPTIONS]'
    echo 'Options:'
    echo '-a  <The AMI ID to use>'
    echo '-i  <The instance type; e.g. t2.micro, m4.xlarge, etc>'
    # echo '-kp <The path to private key file; matching public key should exist; use if no /.aws/credentials exist>'
    echo '-m  <Number of manager nodes>'
    echo '-p  <A prefix for each node name>'
    echo '-r  <The AWS region to use>'
    echo '-rp <Port number for local registry; used when running local image>'
    echo '-s  <The AWS VPC subnet ID>'
    echo '-sg <The AWS VPC security group name>'
    echo '-su <Default SSH user for ami>'
    echo '-v  <The VPC ID to launch the instances in>'
    echo '-w  <Number of agent nodes>'
    echo '-z  <The AWS zone to use; one of: a,b,c,d,e>'
    exit
}

if [ "$#" -ne 24 ]
then 
    usage
fi 

SSH_USER="centos"
REG_PORT=5000
while [ "$1" != "" ]; do
case $1 in
    -a )    shift 
            AMI=$1
            ;;
    -i )    shift
            INSTANCE=$1
            ;;
    # -kp )   shift
    #         KEY_NAME=$1
    #         ;;
    -m )    shift
            MANAGERS=$1
            ;;
    -p )    shift
            NODE_PREFIX=$1
            ;;
    -r )    shift
            REGION=$1
            ;;
    -rp )   shift
            REG_PORT=$1
            ;;
    -s )    shift
            SUBNET=$1
            ;;
    -sg )   shift
            SECURITY_GROUP=$1
            ;;
    -su )   shift
            SSH_USER=$1
            ;;
    -v )    shift
            VPC=$1
            ;;
    -w )    shift
            WORKERS=$1
            ;;
    -z )    shift
            ZONE=$1
            ;;
esac
shift
done

echo "VPC: ${VPC}; SUBNET: ${SUBNET}; REGION: ${REGION}; ZONE: ${ZONE}; AMI: ${AMI}; SECURITY-GROUP: ${SECURITY_GROUP}; INSTANCE-TYPE: ${INSTANCE}; SSH-USER: ${SSH_USER}"
echo "NODE-PREFIX: ${NODE_PREFIX}; NUM-MANAGERS: ${MANAGERS}; NUM-WORKERS: ${WORKERS}"
echo "REGISTRY-PORT: ${REG_PORT}"

# init manager(s)
for number in `seq 1 ${MANAGERS}`
do
echo "Creating machine ${NODE_PREFIX}-m${number}"
docker-machine create --driver amazonec2 \
    --amazonec2-ami ${AMI} \
    --amazonec2-vpc-id ${VPC} \
    --amazonec2-region ${REGION} \
    --amazonec2-zone ${ZONE} \
    --amazonec2-instance-type ${INSTANCE} \
    --amazonec2-subnet-id ${SUBNET} \
    --amazonec2-ssh-user ${SSH_USER} \
    --amazonec2-security-group ${SECURITY_GROUP} ${NODE_PREFIX}"-m"${number}
docker-machine ssh ${NODE_PREFIX}"-m"${number} sudo usermod -a -G docker ${SSH_USER}
done

# init worker(s)
for number in `seq 1 ${WORKERS}`
do
echo "Creating machine ${NODE_PREFIX}-w${number}"
docker-machine create --driver amazonec2 \
    --amazonec2-ami ${AMI} \
    --amazonec2-vpc-id ${VPC} \
    --amazonec2-region ${REGION} \
    --amazonec2-zone ${ZONE} \
    --amazonec2-instance-type ${INSTANCE} \
    --amazonec2-subnet-id ${SUBNET} \
    --amazonec2-ssh-user ${SSH_USER} \
    --amazonec2-security-group ${SECURITY_GROUP} ${NODE_PREFIX}"-w"${number}
docker-machine ssh ${NODE_PREFIX}"-w"${number} sudo usermod -a -G docker ${SSH_USER}
done

# init swarm
MANAGER_PUB_IP=$(docker-machine inspect ${NODE_PREFIX}"-m1" --format='{{.Driver.IPAddress}}')
MANAGER_PRIV_IP=$(docker-machine inspect ${NODE_PREFIX}"-m1" --format='{{.Driver.PrivateIPAddress}}')
echo "Manager Public IP is ${MANAGER_PUB_IP}"
echo "Manager Private IP is ${MANAGER_PRIV_IP}"

# Initialize swarm on the first manager node
docker-machine ssh ${NODE_PREFIX}"-m1" docker swarm init --advertise-addr ${MANAGER_PRIV_IP}
MANAGER_TOKEN=$(docker-machine ssh ${NODE_PREFIX}"-m1" docker swarm join-token manager | sed -n -e 's/^.*--token//p')
WORKER_TOKEN=$(docker-machine ssh ${NODE_PREFIX}"-m1" docker swarm join-token worker | sed -n -e 's/^.*--token//p')

echo "Swarm Manager token is ${MANAGER_TOKEN}"
echo "Swarm Worker token is ${WORKER_TOKEN}"

# configure security groups
# aws ec2 describe-security-groups --filter "Name=group-name,Values=$SECURITY_GROUP"
SECURITY_GROUP_ID=$(aws ec2 describe-security-groups --filter "Name=group-name,Values=${SECURITY_GROUP}" --query SecurityGroups[0].GroupId | sed 's/\"//g')
echo "Security Group ID is ${SECURITY_GROUP_ID}"

# Try updating security group ingress ports
out=$(aws ec2 authorize-security-group-ingress --group-id ${SECURITY_GROUP_ID} --protocol tcp --port 2377 --source-group ${SECURITY_GROUP_ID})
out=$(aws ec2 authorize-security-group-ingress --group-id ${SECURITY_GROUP_ID} --protocol tcp --port 7946 --source-group ${SECURITY_GROUP_ID})
out=$(aws ec2 authorize-security-group-ingress --group-id ${SECURITY_GROUP_ID} --protocol udp --port 7946 --source-group ${SECURITY_GROUP_ID})
out=$(aws ec2 authorize-security-group-ingress --group-id ${SECURITY_GROUP_ID} --protocol tcp --port 4789 --source-group ${SECURITY_GROUP_ID})
out=$(aws ec2 authorize-security-group-ingress --group-id ${SECURITY_GROUP_ID} --protocol udp --port 4789 --source-group ${SECURITY_GROUP_ID})

# Join all managers to the swarm
if test ${MANAGERS} -gt 1
then 
    for number in `seq 2 ${MANAGERS}`
    do 
        echo "Joining ${NODE_PREFIX}-m${number} to swarm as a manager"
        docker-machine ssh ${NODE_PREFIX}"-m"${number} docker swarm join --token ${MANAGER_TOKEN}
    done
fi

# join all workers to the swarm
for number in `seq 1 ${WORKERS}`
do 
    echo "Joining ${NODE_PREFIX}-w${number} to swarm as a worker"
    docker-machine ssh ${NODE_PREFIX}"-w"${number} docker swarm join --token ${WORKER_TOKEN}
done

# Install the registry on the manager, make it available to the nodes as a service
docker-machine ssh ${NODE_PREFIX}"-m1" docker service create --name registry --publish published=${REG_PORT},target=${REG_PORT} registry:2
# if [-z ${REG_PORT}]; then
# eval "$(docker-machine env ${NODE_PREFIZ}"-m1")"

# Default registry
#docker service create --name registry --publish published=5000,target=5000 registry:2

# Change published port
#docker service create --name registry --publish published=5001,target=5000 registry:2

# Change port registry listens on
#docker service create --name registry --env REGISTRY_HTTP_ADDR=0.0.0.0:5001 --publish published=5001,target=5001 registry:2
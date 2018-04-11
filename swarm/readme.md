# Generating An AWS Cluster
The following file is provided for ease of generating an AWS Docker Cluster to be used with the seed cli. This script utilizes docker-machine to do the following:
- create the requested number of manager and worker nodes, 
- initialize a docker swarm and connect the created manager/worker nodes, 
- initialize a registry service on the swarm (using a provided registry url or defaulting to a local registry). This same registry url should be provided to the Seed CLI.
- mount an NFS on the cluster and ensure all machines have access to it

## Script arguments
-a  : The AMI ID to use
-m  : Number of manager nodes
-p  : A prefix for each node name
-r  : The AWS region to use
-s  : The AWS VPC subnet ID
-v  : The VPC ID to launch the instances in
-w  : Number of agent nodes
-z  : The AWS zone to use; one of: a,b,c,d,e
-kp : The path to private key file; matching public key should exist; use if no /.aws/credentials exist
-rp : Port number for local registry; used when running local image
-sg : The AWS VPC security group name
-su : Default SSH user for ami

## Machine Names
Machines are created using the `docker-machine create` command. Machines are named using the following convention:

{Node_prefix}_manager_{number}
{Node_prefix}_worker_{number}

For example: When arguments -p `AWS_NODE`, -m 1, and -w 3 is given, the script will create the following machines:
    AWS_NODE_manager_1
    AWS_NODE_worker_1
    AWS_NODE_worker_2
    AWS_NODE_worker_3

Machines may be displayed using the following command: `docker-machine ls`
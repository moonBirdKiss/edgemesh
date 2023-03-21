#!/bin/bash

set -e

# 错误处理函数
handle_error() {
    echo " * * * * * An error occurred. Exiting... * * * * *  * * * * *  * * * * *  * * * * *  * * * * * "
    exit 1
}
trap 'handle_error' ERR

EDGE0="192.168.1.13"
EDGE1="192.168.1.27"
EDGE2="192.168.1.31"

MASTER="192.168.1.207"
SLAVE1="192.168.1.208"
SLAVE2="192.168.1.209"

# k8s-0上更新
ssh root@$MASTER "ctr -n k8s.io image import /home/ubuntu/Documents/image/edgemesh-amd64.tar"

# k8s-1上更新
ssh root@$SLAVE1 "ctr -n k8s.io image import /home/ubuntu/Documents/image/edgemesh-amd64.tar"

# k8s-2上更新
ssh root@$SLAVE2 "ctr -n k8s.io image import /home/ubuntu/Documents/image/edgemesh-amd64.tar"

# edge-0上更新
 ssh dqf@$EDGE0 "sudo docker load < /home/dqf/image/edgemesh-arm64.tar; sudo docker tag langzijiangnan/edgemesh-agent:v1-linux-arm64 langzijiangnan/edgemesh-agent:v1"

# edge-1上更新
 ssh dqf@$EDGE1 "sudo docker load < /home/dqf/image/edgemesh-arm64.tar; sudo docker tag langzijiangnan/edgemesh-agent:v1-linux-arm64 langzijiangnan/edgemesh-agent:v1"

# edge-2上更新
 ssh dqf@$EDGE2 "sudo docker load < /home/dqf/image/edgemesh-arm64.tar; sudo docker tag langzijiangnan/edgemesh-agent:v1-linux-arm64 langzijiangnan/edgemesh-agent:v1"

# 然后在全局进行更新
ssh root@$MASTER "kubectl delete -f /home/ubuntu/Documents/edgemesh/build/agent/resources; sleep 10; kubectl apply -f /home/ubuntu/Documents/edgemesh/build/agent/resources"
echo "* * * * *  * * * * * * * * Cluster updated * * * * * * * * * * * * * "

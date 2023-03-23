#!/bin/bash



EDGE0="192.168.1.13"
EDGE1="192.168.1.27"
EDGE2="192.168.1.31"

MASTER="192.168.1.207"
SLAVE1="192.168.1.208"
SLAVE2="192.168.1.209"


# k8s-0上清理image
ssh root@$MASTER 'bash -c "crictl rmi $(crictl image ls | grep none | awk '\''{print $3}'\'')"'
echo "* * * * * * * * * MASTER Cleared * * * * * * * * * *"
echo ""

# k8s-1上清理image
ssh root@$SLAVE1 'bash -c "crictl rmi $( crictl image ls | grep none | awk '\''{print $3}'\'')"'
echo "* * * * * * * * * SLAVE1 Cleared * * * * * * * * * *"
echo ""

# k8s-2上清理image
ssh root@$SLAVE2 'bash -c "crictl rmi $( crictl image ls | grep none | awk '\''{print $3}'\'')"'
echo "* * * * * * * * * SLAVE2 Cleared * * * * * * * * * *"
echo ""

# edge-0上清理image
ssh dqf@$EDGE0 'bash -c "docker rmi $(docker image ls | grep none | awk '\''{print $3}'\'')"'
echo "* * * * * * * * * EDGE0 Cleared * * * * * * * * * *"
echo ""

# docker rmi $(docker image ls | grep none | awk '{print $3}')

# edge-1上清理image
ssh dqf@$EDGE1 'bash -c "docker rmi $(docker image ls | grep none | awk '\''{print $3}'\'')"'
echo "* * * * * * * * * EDGE1 Cleared * * * * * * * * * *"
echo ""

# edge-2上清理image
 ssh dqf@$EDGE2 'bash -c "docker rmi $(docker image ls | grep none | awk '\''{print $3}'\'')"'
echo "* * * * * * * * * EDGE2 Cleared * * * * * * * * * *"

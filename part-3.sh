#!/bin/bash



EDGE4="192.168.1.61"
EDGE5="192.168.1.62"
EDGE6="192.168.1.64"

MASTER="192.168.1.70"
SLAVE1="192.168.1.71"
#SLAVE2="192.168.1.209"

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
ssh dqf@$EDGE4 'bash -c "docker rmi $(docker image ls | grep none | awk '\''{print $3}'\'')"'
echo "* * * * * * * * * EDGE4 Cleared * * * * * * * * * *"
echo ""

# docker rmi $(docker image ls | grep none | awk '{print $3}')

# edge-1上清理image
ssh dqf@$EDGE5 'bash -c "docker rmi $(docker image ls | grep none | awk '\''{print $3}'\'')"'
echo "* * * * * * * * * EDGE5 Cleared * * * * * * * * * *"
echo ""

# edge-2上清理image
 ssh dqf@$EDGE6 'bash -c "docker rmi $(docker image ls | grep none | awk '\''{print $3}'\'')"'
echo "* * * * * * * * * EDGE6 Cleared * * * * * * * * * *"

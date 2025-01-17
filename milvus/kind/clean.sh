docker image ls |grep -v TAG |grep -v kind | awk '{printf "%s:%s\n", $1, $2}'  |  xargs -I {} docker rmi {}

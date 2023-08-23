docker stop python-fs-proxy
docker rm python-fs-proxy
docker run --network proxy-net --name python-fs-proxy -itd goxii:latest ./Goxii
docker connect internal python-fs-proxy
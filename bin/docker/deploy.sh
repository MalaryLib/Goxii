docker container stop goxii-cont
docker container rm goxii-cont

go build

docker build -t goxii .
docker run -itd --network proxy-net -p 8080:8080 --name goxii-cont goxii
docker network connect synonym-net goxii-cont

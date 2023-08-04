<img src="https://repository-images.githubusercontent.com/674188947/f437df51-622e-496b-8184-45bc27529dec"/>

# Goxii: TCP Layer Go Proxy 

This 

> This is still in active development, for more information... wait like two days. 

### Want to Run?
First build the image using docker, if you don't have it download it. From the Goxii source folder run...

```docker
docker build -t goxii-img .
```

Now run a container 

```
sh deploy.sh
```

Now jump in the container (you can also append a CMD sh start.sh) in the Dockerfile, I haven't done this for dev reasons.

```
docker exec -it goxii-cont bash
```

Now that you are in the goxii-cont container, we start the proxy

```
sh start.sh
```

### But it's not connecting to anything...
Yeah, I've got it hard-coded to a webserver runnning on an independent docker network. That info is hardcoded in both the `deploy.sh` file and the `start.sh` file. 

More on this later.

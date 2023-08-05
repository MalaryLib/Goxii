<img src="https://repository-images.githubusercontent.com/674188947/2715362d-44ef-4e33-98dd-8f7a617fa168"/>

#

```bash
> goxii 8080 172.20.0.2 443
```

### Using <mark>Goxii</mark>

#### Pre-Requisites
Goxii is designed as a drop-in proxy for securing docker containers that host internal data and or information. _Before using Goxii, please install the 
following software:_
- [ ] [Docker](https://docs.docker.com/get-docker/)
- [ ] [Docker Compose](https://docs.docker.com/compose/install/)
- [ ] A container to test with!

#### Now what?

Using Goxii is pretty straight forward. You can simply clone this repository:
```bash
git clone https://github.com/MalaryLib/Goxii.git
```

##### Configuring Goxii

> Because this is a container, **all of the configuration is first done in the [`compose.yml`](/compose.yaml) file located in this repository.**

| Options | How to configure | What does this change? |
|--|--|--|
| Destination IP | [`compose.yml`](/compose.yaml) | This is where your destination container is located. Must be an IP address as of Goxii-v1.0.
| Destination Port | [`compose.yml`](/compose.yaml) | This is the port that your destination container is listening on. |
| External-Facing Proxy Network | [`compose.yml`](/compose.yaml) | This is the network that you are expecting to potentially be reachable from the outside. Add it to the networks and set it as external.
| Port Goxii Listens On | [`compose.yml`](/compose.yaml) | Goxii listens on port 8081 by default, change this in the `ports` section of the [`compose.yml`](/compose.yaml) file.
| Allowed IPs | [/resources/.ips](/resources/.ips) | Goxii reads the ips in this file on start-up to get a list of allowed IPs. This is a temporary bug in v1.0. Future versions will use a token based authentication scheme.

You can change some other things too. For example, you can run the goxii binary with the following parameters:

```bash
goxii <port to listen to> <Destination IP> <Destination Port>
```



##### Containerizing Goxii

We now start the container with goxii! You most likely have to run this as root or with sudo for socket priviledges.

> _These commands are run from the Goxii folder_, if you're running with root omit the sudo commands.

```bash
sudo docker compose up -d
```

* This should run without any issues. Ensure you are connected to the internet and have storage, etc.

From here you are ready to start curl-ing, rest-ing, or whatever-ing your private container. The logs are pretty descriptive.
# container_tools installer
This is simple installer container for container tools.

Container tools provide docker_rdma_sriov tool for running docker containers with RDMA/DPDK sriov devices.

## How to install tools?
```
docker run --net=host -v /usr/bin:/tmp mellanox/container_tools_installer
```
This install following container tools.
1. docker_rdma_sriov

## How to use this tools?

### How to run docker container using sriov-plugin?

**1** Run the sriov-plugin.
```
docker run -v /run/docker/plugins:/run/docker/plugins -v /etc/docker:/etc/docker --net=host --privileged mellanox/sriov-plugin
```
**2** Create sriov based network
```
docker network create -d sriov --subnet=194.168.1.0/24 -o netdevice=ens2f0 -o mode=sriov mynet
```
**3** Start container which uses rdma/dpdk devices
```
docker_rdma_sriov run --net=mynet -it centos bash
```

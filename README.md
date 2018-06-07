## container_tools installer
Few useful container orchestration, deployment tools when using RDMA

Container tools provide docker_rdma_sriov tool for managing and docker containers for
RDMA and DPDK application uses using SRIOV networking plugin.

This is simple installer container to for these helpful container tools.

## How to install tools?
```
docker run --net=host -v /usr/bin:/tmp mellanox/container_tools
```
This install following container tools.
1. docker_rdma_sriov

## How to use this tools?

# How to run docker container using sriov-plugin?

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

# Enable sriov for a PF netdevice
This command enables SRIOV for netdevice ib0 and does necessary configuration.
```
docker_rdma_sriov sriov enable --netdev=ib0
```

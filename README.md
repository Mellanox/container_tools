# container_tools installer
This is simple installer container for container tools.

Container tools provide docker_rdma_sriov tool for running docker containers with RDMA/DPDK sriov devices.

## How to install tools?
```
docker run --net=host -v /usr/bin:/tmp rdma/container_tools_installer
```
This install following container tools.
1. docker_rdma_sriov

## How to use this tools?

### How to run docker container using sriov-plugin?

**1** Run the sriov-plugin.
```
docker run -v /run/docker/plugins:/run/docker/plugins -v /etc/docker:/etc/docker -v /var/run:/var/run --net=host --privileged rdma/sriov-plugin
```
**2** Create sriov based network
```
docker network create -d sriov --subnet=194.168.1.0/24 -o netdevice=ens2f0 -o mode=sriov mynet
```
**3** Start container which uses rdma/dpdk devices
```
docker_rdma_sriov run --net=mynet -it centos bash
```

### How to run docker container without a sriov-plugin?

**1** Set RDMA subsystem in exclusive mode.
```
docker_rdma_sriov rdmanetns set --mode=exclusive
```
This will be supported in recent kernel such as kernel 5.2 or MOFED 4.7 or higher.
If any net namespace is created by user, this command will fail.
User must make sure that no containers are running or no net namespaces exist in the system.

**2** Configure sriov and if necessary switchdev mode.
```
docker_rdma_sriov sriov enable -n ens2f0
```

**3** Run dummy container in new net namespace without a plugin
```
docker run --net=none -d mellanox/sleepyhead
```

Get the container id to use it later for netdevice configuration, say 8d6cb8f49507.

**4** Provision one VF say vf index=1 to a dummy container.

Assign VF 1 netdev of PF PCI netdevice ens2f0, name as eth0 in container.

```
docker_rdma_sriov run sriov attachndev --container 8d6cb8f49507 --netdev ens2f0 --vf 1 -N eth0
```

Assign VF 1 RDMA device of PF PCI netdevice ens2f0, in container.
This is supported only in exclusive mode.

```
docker_rdma_sriov sriov attachrdev -container 8d6cb8f49507 --netdev ens2f0 -vf 1
```

**5** Assign IP address and gateway address to this VF netdevice

```
docker_rdma_sriov net ipcfg -i 194.168.1.1/24 -g 194.168.1.45 -n eth1 -c 8d6cb8f49507
```

**6** Now start a real RDMA application container

```
docker_rdma_sriov run -it --net=container:8d6cb8f49507 mellanox/mlnx_ofed_linux-4.4-2.0.7.0-centos7.4 bash
```

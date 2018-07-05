package main

import (
	"context"
	"fmt"
	"github.com/Mellanox/rdmamap"
	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
	"os"
)

func DumpAllContainersRdmaStats() {
	cli, err := getRightClient()
	if err != nil {
		return
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		fmt.Println("Fail to get container list", err)
		return
	}

	for _, container := range containers {
		//There is no need to get stats of the containers running
		//in default host network namespace.
		//It can be queries directly via sysfs access or
		//future rdmatool.
		if container.HostConfig.NetworkMode == "host" {
			continue
		}
		fmt.Println("Container = ", container.ID)
		fmt.Println("State = ", container.State)
		fmt.Println("Status = ", container.Status)
		fmt.Println("Network mode = ", container.HostConfig.NetworkMode)

		rdmamap.GetDockerContainerRdmaStats(container.ID)
	}
}

func execUserStatsCmd(userCmdArgs []string, args []string) {

	if len(args) > 0 {
		for _, cid := range args {
			rdmamap.GetDockerContainerRdmaStats(cid)
		}
	} else {
		DumpAllContainersRdmaStats()
	}
}

func execStatsCmd(cmd *cobra.Command, args []string) {
	if len(os.Args) <= 2 {
		fmt.Printf("Show rdma statistics of container\n\n")
		fmt.Printf("Usage:\n")
		fmt.Printf("docker_rdma_sriov stats CONTAINER_ID\n")
		return
	}
	execUserStatsCmd(os.Args[2:], args)
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "RDMA Statistics of a container",
	Run:   execStatsCmd,
}

func init() {

}

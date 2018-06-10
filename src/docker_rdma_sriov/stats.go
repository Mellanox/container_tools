package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"log"
	"net"
	"github.com/vishvananda/netns"
	"github.com/Mellanox/rdmamap"
)

func printRdmaStats(device string, stats *rdmamap.RdmaStats) {

	for _, portstats := range stats.PortStats {
		fmt.Printf("device: %s, port: %d\n", device, portstats.Port)
		fmt.Println("Hw stats:")
		for _, entry := range portstats.HwStats {
			fmt.Printf("%s = %d\n", entry.Name, entry.Value)
		}
		fmt.Println("Stats:")
		for _, entry := range portstats.HwStats {
			fmt.Printf("%s = %d\n", entry.Name, entry.Value)
		}
	}
}

func execUserStatsCmd(userCmdArgs []string, args []string) {

	fmt.Println("args = ", args)

	originalHandle, err := netns.Get()
	if err != nil {
		log.Println("Fail to get handle of current net ns", err)
		return
	}
	nsHandle, err := netns.GetFromDocker(args[0])
	if err != nil {
		log.Println("Invalid docker id: ", args[0])
		return
	}
	log.Println("nsHandle = ", nsHandle)
	netns.Set(nsHandle)
	
	ifaces, err := net.Interfaces()
	if err != nil {
		netns.Set(originalHandle)
		return
	}
	fmt.Printf("Interfaces: %v\n", ifaces)
	for _, iface := range ifaces {
		if iface.Name == "lo" {
			continue
		}
		rdmadev, err := rdmamap.GetRdmaDeviceForNetdevice(iface.Name)
		if err != nil {
			continue
		}
		rdmastats, err := rdmamap.GetRDmaSysfsAllPortsStats(rdmadev)
		if err != nil {
			log.Println("Fail to query device stats: ", err)
			continue
		}
		log.Println("rdma device = ", rdmadev)
		printRdmaStats(rdmadev, &rdmastats)
	}

	netns.Set(originalHandle)
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
	/*
	vfFlags := statsCmd.Flags()
	vfFlags.StringVarP(&vfUserArg, "vf", "n", "0", "vf index")
	*/
}

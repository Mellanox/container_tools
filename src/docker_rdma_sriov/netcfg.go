package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"runtime"
)

var netcfgCmds = &cobra.Command{
	Use:   "net",
	Short: "Network device, IP, gateway configuration for netdevice in container",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return nil
	},
}

var netdev string
var ipAddrMask string
var gwAddr string

func listNdevNetcfgFunc(cmd *cobra.Command, args []string) {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	nsHandle, err := netns.GetFromDocker(containerId)
	if err != nil {
		fmt.Println("Invalid container id: ", containerId)
		return
	}
	originalHandle, err := netns.Get()
	if err != nil {
		fmt.Println("Fail to get handle of current net ns", err)
		return
	}
	netns.Set(nsHandle)

	ifaces, err := net.Interfaces()
	if err != nil {
		netns.Set(originalHandle)
		return
	}
	fmt.Printf("Net Interfaces: \n", ifaces)
	for _, iface := range ifaces {
		fmt.Printf("%v\n", iface)
	}
	fmt.Printf("\n")

	netns.Set(originalHandle)
}

func ipcfg(cid, netdev string, ip string, gw string) error {
	//var gwNet *net.IPNet
	var gwIP net.IP

	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	nsHandle, err := netns.GetFromDocker(containerId)
	if err != nil {
		fmt.Println("Invalid container id: ", containerId)
		return err
	}
	ipv4, ipv4Net, err := net.ParseCIDR(ip)
	if err != nil {
		fmt.Printf("Invalid IP addr, err=%v\n", ip, err)
		return err
	}
	var address = &net.IPNet{IP: ipv4, Mask: ipv4Net.Mask}
	var addr = &netlink.Addr{IPNet: address}

	if len(gw) != 0 {
		gwIP, _, err = net.ParseCIDR(ip)
		if err != nil {
			fmt.Printf("Invalid gateway address, err=%v\n", gw)
			return err
		}
	}
	originalHandle, err := netns.Get()
	if err != nil {
		fmt.Println("Fail to get handle of current net ns", err)
		return err
	}
	netns.Set(nsHandle)
	link, err := netlink.LinkByName(netdev)
	if err != nil {
		fmt.Printf("Netdev %v not found for err=%v\n", netdev, err)
		netns.Set(originalHandle)
		return err
	}

	err = netlink.AddrAdd(link, addr)
	if err != nil {
		fmt.Printf("Fail to add IP address err=%v\n", err)
		netns.Set(originalHandle)
		return err

	}
	err = netlink.LinkSetUp(link)
	if err != nil {
		fmt.Printf("Netdev %v fail to bringup err=%v\n", netdev, err)
		netns.Set(originalHandle)
		return err
	}

	if len(gw) != 0 {
		defaultRoute := netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       nil,
			Gw:        gwIP,
		}
		err = netlink.RouteAdd(&defaultRoute)
		if err != nil {
			fmt.Printf("Fail to set default gateway err=%v\n", err)
		}
	}
	netns.Set(originalHandle)
	return nil
}

func ipcfgNetcfgFunc(cmd *cobra.Command, args []string) {
	if netdev == "" {
		fmt.Println("Please specify valid netdevice")
		return
	}
	if containerId == "" {
		fmt.Println("Please specify container id")
		return
	}
	ipcfg(containerId, netdev, ipAddrMask, gwAddr)
}

var listNetcfgCmd = &cobra.Command{
	Use:   "list",
	Short: "List netdevices of a container",
	Run:   listNdevNetcfgFunc,
}

var ipcfgNetcfgCmd = &cobra.Command{
	Use:   "ipcfg",
	Short: "Configure ip, gateway of the netdevice in a container",
	Run:   ipcfgNetcfgFunc,
}

func init() {
	listFlags := listNetcfgCmd.Flags()
	listFlags.StringVarP(&containerId, "container", "c", "", "Container id whose netdevices to list")

	ipcfgFlags := ipcfgNetcfgCmd.Flags()
	ipcfgFlags.StringVarP(&netdev, "netdev", "n", "", "PF netdevice whose VF to attach to container")
	ipcfgFlags.StringVarP(&containerId, "container", "c", "", "Container id to attach to")
	ipcfgFlags.StringVarP(&ipAddrMask, "ipaddr", "i", "", "IPv4 address in format 194.168.1.1/24")
	ipcfgFlags.StringVarP(&gwAddr, "gateway", "g", "", "Gateway address in format 194.168.1.1")
}

/* add new sriov command here */
var netcfgList = [...]*cobra.Command{
	listNetcfgCmd,
	ipcfgNetcfgCmd,
}

func init() {
	for _, cmds := range netcfgList {
		netcfgCmds.AddCommand(cmds)
	}
}

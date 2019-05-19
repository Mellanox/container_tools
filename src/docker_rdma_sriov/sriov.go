package main

import (
	"context"
	"fmt"
	"github.com/Mellanox/sriovnet"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"syscall"
)

var sriovCmds = &cobra.Command{
	Use:   "sriov",
	Short: "sriov management commands for netdevices",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return nil
	},
}

var pfNetdev string
var vfIndex int
var containerId string
var vfNewNdevName string

func enableSriovFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specify valid PF netdevice")
		return
	}

	err1 := sriovnet.EnableSriov(pfNetdev)
	if err1 != nil {
		return
	}

	handle, err2 := sriovnet.GetPfNetdevHandle(pfNetdev)
	if err2 != nil {
		return
	}
	err3 := sriovnet.ConfigVfs(handle, false)
	if err3 != nil {
		return
	}
}

func disableSriovFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specific valid PF netdevice")
		return
	}
	sriovnet.DisableSriov(pfNetdev)
}

func listSriovFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specify valid PF netdevice")
		return
	}
	handle, err2 := sriovnet.GetPfNetdevHandle(pfNetdev)
	if err2 != nil {
		return
	}

	for _, vf := range handle.List {
		vfName := sriovnet.GetVfNetdevName(handle, vf)
		fmt.Printf("%v ", vfName)
	}
	fmt.Printf("\n")
}

func unbindSriovFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specify valid PF netdevice")
		return
	}
	handle, err2 := sriovnet.GetPfNetdevHandle(pfNetdev)
	if err2 != nil {
		return
	}

	if vfIndex != -1 {
		var found bool
		var err error
		for _, vf := range handle.List {
			if vfIndex == vf.Index {
				found = true
				fmt.Printf("Unbinding VF: %d\n", vf.Index)
				err = sriovnet.UnbindVf(handle, vf)
				if err != nil {
					fmt.Printf("Fail to Unbind VF: ", err)
					break
				}
			}
		}
		if found == false {
			fmt.Println("VF index = %d not found\n", vfIndex)
		}
	} else {
		for _, vf := range handle.List {
			fmt.Printf("Unbinding VF: %d\n", vf.Index)
			err := sriovnet.UnbindVf(handle, vf)
			if err != nil {
				fmt.Println("Fail to unbind VF: ", err)
				fmt.Printf("Continu to bind other VFs\n")
			}
		}
	}

	for _, vf := range handle.List {
		vfName := sriovnet.GetVfNetdevName(handle, vf)
		fmt.Printf("%v ", vfName)
	}
	fmt.Printf("\n")
}

func BindVf(pfNetdev string, handle *sriovnet.PfNetdevHandle, vf *sriovnet.VfObj) error {
	fmt.Printf("Binding VF: %d\n", vf.Index)
	err := sriovnet.BindVf(handle, vf)
	if err != nil {
		fmt.Println("Fail to bind VF: ", err)
		return nil
	}
	mode, _ := GetDevlinkMode(pfNetdev)
	if mode != "switchdev" {
		fmt.Println("Skipping VF rep link config")
	}
	err = SetVfRepresentorLinkUp(pfNetdev, vf.Index)
	if err != nil {
		fmt.Println("Fail to bind VF: ", err)
	}
	return err
}

func bindSriovFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specify valid PF netdevice")
		return
	}
	handle, err2 := sriovnet.GetPfNetdevHandle(pfNetdev)
	if err2 != nil {
		return
	}

	if vfIndex != -1 {
		var found bool
		for _, vf := range handle.List {
			if vfIndex != vf.Index {
				continue
			}
			found = true
			err := BindVf(pfNetdev, handle, vf)
			if err != nil {
				break
			}
		}
		if found == false {
			fmt.Println("VF index = %d not found\n", vfIndex)
		}
	} else {
		for _, vf := range handle.List {
			err := BindVf(pfNetdev, handle, vf)
			if err != nil {
				break
			}
		}
	}

	for _, vf := range handle.List {
		vfName := sriovnet.GetVfNetdevName(handle, vf)
		fmt.Printf("%v ", vfName)
	}
	fmt.Printf("\n")
}

func getSandboxKeyFd(cid string) (int, error) {

	cli, err := getRightClient()
	if err != nil {
		fmt.Printf("Fail to get docker client info, err = %v\n", err)
		return -1, err
	}
	info, err := cli.ContainerInspect(context.Background(), cid)
	if err != nil {
		fmt.Printf("Fail to get container info, err = %v\n", err)
		return -1, err
	}
	fd, err := syscall.Open(info.NetworkSettings.SandboxKey, syscall.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("Fail to open sandbox fd %v, err=%v\n", info.NetworkSettings.SandboxKey, err)
		return -1, err
	}
	return fd, nil
}

func renameNetdevName(cid, newname string, oldname string) error {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	originalHandle, err := netns.Get()
	if err != nil {
		fmt.Println("Fail to get handle of current net ns", err)
		return err
	}

	nsHandle, err := netns.GetFromDocker(containerId)
	if err != nil {
		fmt.Println("Invalid docker id: ", containerId)
		return err
	}
	netns.Set(nsHandle)
	vfLink, err := netlink.LinkByName(oldname)
	if err != nil {
		fmt.Printf("Netdev not found for vf index = %d err=%v\n", vfIndex, err)
		return err
	}
	err = netlink.LinkSetName(vfLink, newname)
	if err != nil {
		fmt.Printf("Netdev %v not found\n", oldname)
	}
	netns.Set(originalHandle)
	return nil
}

func attachNdevSriovFunc(cmd *cobra.Command, args []string) {
	var found bool
	var vf *sriovnet.VfObj

	if pfNetdev == "" {
		fmt.Println("Please specify valid PF netdevice")
		return
	}
	handle, err := sriovnet.GetPfNetdevHandle(pfNetdev)
	if err != nil {
		fmt.Printf("Fail to get handle, err=%v\n", err)
		return
	}

	if vfIndex == -1 {
		fmt.Println("Please specific valid VF index")
		return
	}
	for _, vf = range handle.List {
		if vfIndex == vf.Index {
			found = true
			break
		}
	}
	if found == false {
		fmt.Printf("VF index = %d not found\n", vfIndex)
		return
	}
	if containerId == "" {
		fmt.Println("Please specify container id")
		return
	}
	vfNetdev := sriovnet.GetVfNetdevName(handle, vf)
	if vfNetdev == "" {
		fmt.Printf("Netdev not found for vf index = %d\n", vfIndex)
		return
	}
	vfLink, err := netlink.LinkByName(vfNetdev)
	if err != nil {
		fmt.Printf("Netdev not found for vf index = %d err=%v\n", vfIndex, err)
		return
	}

	/* inspect container and get sandbox key handle */
	sandboxkeyFd, err := getSandboxKeyFd(containerId)
	if err != nil {
		return
	}
	err = netlink.LinkSetNsFd(vfLink, sandboxkeyFd)
	if err != nil {
		fmt.Printf("Fail to move vf Index %d, netdev=%v to container, err=%v\n",
			vfIndex, vfNetdev, err)
	}
	renameNetdevName(containerId, vfNewNdevName, vfNetdev)
}

var enableSriovCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable sriov for PF netdevice",
	Run:   enableSriovFunc,
}

var disableSriovCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable sriov for PF netdevice",
	Run:   disableSriovFunc,
}

var listSriovCmd = &cobra.Command{
	Use:   "list",
	Short: "List sriov netdevices for PF netdevice",
	Run:   listSriovFunc,
}

var unbindSriovCmd = &cobra.Command{
	Use:   "unbind",
	Short: "Unbind a specific or all VFs of a PF netdevice",
	Run:   unbindSriovFunc,
}

var bindSriovCmd = &cobra.Command{
	Use:   "bind",
	Short: "bind a specific or all VFs of a PF netdevice",
	Run:   bindSriovFunc,
}

var attachNdevSriovCmd = &cobra.Command{
	Use:   "attachndev",
	Short: "Attach VF's netdevice to an existing container specified using containerid",
	Run:   attachNdevSriovFunc,
}

func init() {
	enableFlags := enableSriovCmd.Flags()
	enableFlags.StringVarP(&pfNetdev, "netdev", "n", "", "enable sriov for the PF netdevice")

	disableFlags := disableSriovCmd.Flags()
	disableFlags.StringVarP(&pfNetdev, "netdev", "n", "", "disable sriov for the PF netdevice")

	listFlags := listSriovCmd.Flags()
	listFlags.StringVarP(&pfNetdev, "netdev", "n", "", "List netdevices of the PF netdevice")

	unbindFlags := unbindSriovCmd.Flags()
	unbindFlags.IntVarP(&vfIndex, "vf", "v", -1, "vf index to unbind")
	unbindFlags.StringVarP(&pfNetdev, "netdev", "n", "", "PF netdevice whose VFs to unbind")

	bindFlags := bindSriovCmd.Flags()
	bindFlags.IntVarP(&vfIndex, "vf", "v", -1, "vf index to bind")
	bindFlags.StringVarP(&pfNetdev, "netdev", "n", "", "PF netdevice whose VFs to bind")

	attachFlags := attachNdevSriovCmd.Flags()
	attachFlags.IntVarP(&vfIndex, "vf", "v", -1, "vf index to attach to container")
	attachFlags.StringVarP(&pfNetdev, "netdev", "n", "", "PF netdevice whose VF to attach to container")
	attachFlags.StringVarP(&containerId, "container", "c", "", "Container id to attach to")
	attachFlags.StringVarP(&vfNewNdevName, "vfNewNdevName", "N", "eth0", "VF netdev name in container")
}

/* add new sriov command here */
var sriovCmdList = [...]*cobra.Command{
	enableSriovCmd,
	disableSriovCmd,
	listSriovCmd,
	unbindSriovCmd,
	bindSriovCmd,
	attachNdevSriovCmd,
}

func init() {
	for _, cmds := range sriovCmdList {
		sriovCmds.AddCommand(cmds)
	}
}

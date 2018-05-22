package main

import (
	"fmt"
	"github.com/Mellanox/sriovnet"
	"github.com/spf13/cobra"
)

var sriovCmds = &cobra.Command{
	Use:   "sriov",
	Short: "sriov management commands for netdevices",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return nil
	},
}

var enableNetdev string
var disableNetdev string
var listNetdev string

func enableSriovFunc(cmd *cobra.Command, args []string) {
	err1 := sriovnet.EnableSriov(enableNetdev)
	if err1 != nil {
		return
	}

	handle, err2 := sriovnet.GetPfNetdevHandle(enableNetdev)
	if err2 != nil {
		return
	}
	err3 := sriovnet.ConfigVfs(handle, false)
	if err3 != nil {
		return
	}
}

func disableSriovFunc(cmd *cobra.Command, args []string) {
	sriovnet.DisableSriov(disableNetdev)
}

func listSriovFunc(cmd *cobra.Command, args []string) {
	handle, err2 := sriovnet.GetPfNetdevHandle(listNetdev)
	if err2 != nil {
		return
	}

	for _, vf := range handle.List {
		vfName := sriovnet.GetVfNetdevName(handle, vf)
		fmt.Printf("%v ", vfName)
	}
	fmt.Printf("\n")
}

var enableSriovCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable sriov",
	Run:   enableSriovFunc,
}

var disableSriovCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable sriov",
	Run:   disableSriovFunc,
}

var listSriovCmd = &cobra.Command{
	Use:   "list",
	Short: "List sriov",
	Run:   listSriovFunc,
}

func init() {
	enableFlags := enableSriovCmd.Flags()
	enableFlags.StringVarP(&enableNetdev, "netdev", "n", "netdev", "enable sriov for the PF netdevice")

	disableFlags := disableSriovCmd.Flags()
	disableFlags.StringVarP(&disableNetdev, "netdev", "n", "netdev", "disable sriov for the PF netdevice")

	listFlags := listSriovCmd.Flags()
	listFlags.StringVarP(&listNetdev, "netdev", "n", "netdev", "List netdevices of the PF")
}

/* add new sriov command here */
var sriovCmdList = [...]*cobra.Command{
	enableSriovCmd,
	disableSriovCmd,
	listSriovCmd,
}

func init() {
	for _, cmds := range sriovCmdList {
		sriovCmds.AddCommand(cmds)
	}
}

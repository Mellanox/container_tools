package main

import (
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

func init() {
	enableFlags := enableSriovCmd.Flags()
	enableFlags.StringVarP(&enableNetdev, "netdev", "n", "netdev", "netdevice to enable sriov for")

	disableFlags := disableSriovCmd.Flags()
	disableFlags.StringVarP(&disableNetdev, "netdev", "n", "netdev", "netdevice to disable sriov for")
}

/* add new sriov command here */
var sriovCmdList = [...]*cobra.Command{
	enableSriovCmd,
	disableSriovCmd,
}

func init() {
	for _, cmds := range sriovCmdList {
		sriovCmds.AddCommand(cmds)
	}
}

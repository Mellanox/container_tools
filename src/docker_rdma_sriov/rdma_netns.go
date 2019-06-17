package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
)

var mode string

var rdmaNetnsCfgCmds = &cobra.Command{
	Use:   "rdmanetns",
	Short: "RDMA Net namespace configuration commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return nil
	},
}

func showRdmaNetnsFunc(cmd *cobra.Command, args []string) {
	rdmamode, err := netlink.RdmaSystemGetNetnsMode()
	if err != nil {
		fmt.Printf("Fail to query rdma net namespace mode: err = ", err.Error())
		return
	}
	fmt.Printf("RDMA net namespace device sharing mode = %v\n", rdmamode)
}

func setRdmaNetnsFunc(cmd *cobra.Command, args []string) {
	if mode == "" || (mode != "exclusive" && mode != "shared") {
		fmt.Println("Please specify valid net namespace mode")
		return
	}
	err := netlink.RdmaSystemSetNetnsMode(mode)
	if err != nil {
		fmt.Printf("Fail to set rdma net namespace mode: err = %v\n", err.Error())
		return
	}
}

var showRdmaNetnsCfgCmd = &cobra.Command{
	Use:   "show",
	Short: "show RDMA device net namespace sharing mode",
	Run:   showRdmaNetnsFunc,
}

var setRdmaNetnsCfgCmd = &cobra.Command{
	Use:   "set",
	Short: "Configure ip, gateway of the netdevice in a container",
	Run:   setRdmaNetnsFunc,
}

func init() {
	setNetnsModeFlags := setRdmaNetnsCfgCmd.Flags()
	setNetnsModeFlags.StringVarP(&mode, "mode", "m", "", "mode=exclusive or mode=shared")
}

/* add new sriov command here */
var rdmaNetnsList = [...]*cobra.Command{
	showRdmaNetnsCfgCmd,
	setRdmaNetnsCfgCmd,
}

func init() {
	for _, cmds := range rdmaNetnsList {
		rdmaNetnsCfgCmds.AddCommand(cmds)
	}
}

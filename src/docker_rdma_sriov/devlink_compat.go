package main

import (
	"fmt"
	"github.com/Mellanox/sriovnet"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

const (
	devlinkCompatFile       = "compat/devlink/mode"
	devlinkCompatInlineFile = "compat/devlink/inline"
	devlinkCompatEncapFile  = "compat/devlink/encap"
)

var devlinkMode string

var devlinkCmds = &cobra.Command{
	Use:   "devlink",
	Short: "devlink management commands for netdevices",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return nil
	},
}

func GetDevlinkMode(netdev string) (string, error) {
	file := filepath.Join(sriovnet.NetSysDir, netdev, devlinkCompatFile)
	fileObj := fileObject{
		Path: file,
	}

	mode, err := fileObj.Read()
	if err != nil {
		return "", err
	} else {
		return strings.Trim(mode, "\n"), nil
	}
}

func SetDevlinkMode(netdev string, newMode string) error {
	file := filepath.Join(sriovnet.NetSysDir, netdev, devlinkCompatFile)
	fileObj := fileObject{
		Path: file,
	}
	return fileObj.Write(newMode)
}

func SetDevlinkInlineMode(netdev string, newMode string) error {
	file := filepath.Join(sriovnet.NetSysDir, netdev, devlinkCompatInlineFile)
	fileObj := fileObject{
		Path: file,
	}
	return fileObj.Write(newMode)
}

func SetDevlinkEncapMode(netdev string, newMode string) error {
	file := filepath.Join(sriovnet.NetSysDir, netdev, devlinkCompatEncapFile)
	fileObj := fileObject{
		Path: file,
	}
	return fileObj.Write(newMode)
}

func getDevlinkModeFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specific valid PF netdevice")
		return
	}

	currentMode, err := GetDevlinkMode(pfNetdev)
	if err != nil {
		fmt.Printf("Fail to get the devlink mode: \n", err)
		return
	}
	fmt.Println("devlink mode: ", currentMode)
}

func setDevlinkModeFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specific valid PF netdevice")
		return
	}
	if devlinkMode == "" || (devlinkMode != "legacy" && devlinkMode != "switchdev") {
		fmt.Println("Please specific valid devlink mode (legacy/switchdev)")
		return
	}
	err := SetDevlinkEncapMode(pfNetdev, "none")
	if err != nil {
		fmt.Println("Fail to set the encap mode: ", err)
		return
	}

	err = SetDevlinkInlineMode(pfNetdev, "transport")
	if err != nil {
		fmt.Println("Fail to set the devlink inline mode: ", err)
		fmt.Println("Ignoring the inline mode error.")
	}

	if devlinkMode == "switchdev" {
		err = setupSwitchdevUdevRule(pfNetdev)
		if err != nil {
			fmt.Println("Fail to setup udev rules ", err)
			return
		}
	}

	err = SetDevlinkMode(pfNetdev, devlinkMode)
	if err != nil {
		fmt.Println("Fail to set the devlink mode: ", err)
		return
	}
}

var getDevlinkModeCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the devlink mode of the PF netdevice",
	Run:   getDevlinkModeFunc,
}

var setDevlinkModeCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the devlink mode of the PF netdevice",
	Run:   setDevlinkModeFunc,
}

func init() {
	getFlags := getDevlinkModeCmd.Flags()
	getFlags.StringVarP(&pfNetdev, "netdev", "n", "", "PF netdevice")

	setFlags := setDevlinkModeCmd.Flags()
	setFlags.StringVarP(&pfNetdev, "netdev", "n", "", "PF netdevice")
	setFlags.StringVarP(&devlinkMode, "mode", "m", "", "new devlink mode (legacy/switchdev)")
}

/* add new sriov command here */
var devlinkCmdList = [...]*cobra.Command{
	getDevlinkModeCmd,
	setDevlinkModeCmd,
}

func init() {
	for _, cmds := range devlinkCmdList {
		devlinkCmds.AddCommand(cmds)
	}
}

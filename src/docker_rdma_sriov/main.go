package main

import (
	"context"
	"fmt"
	"github.com/Mellanox/rdmamap"
	"github.com/Mellanox/sriovnet"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	appVersion = "0.0.1"
)

func getDockerNetworkResourceForName(networkName string) *types.NetworkResource {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		panic(err)
	}

	for _, network := range networks {
		if network.Name != networkName {
			continue
		}
		return &network
	}
	return nil
}

//Returns name of the network if provided in the docker run command.
func getNonDefaultNetwork(userCmdArgs []string) string {

	for _, item := range userCmdArgs {
		if strings.Contains(item, "--net=") == false {
			continue
		}
		name := strings.Split(item, "=")
		if len(name) < 2 {
			return ""
		}
		return name[1]
	}
	return ""
}

func toCharDevCmdArgs(devices []string) []string {
	var cmds []string

	for _, dev := range devices {
		cmd := "--device=" + dev
		cmds = append(cmds, cmd)
	}
	return cmds
}

func allocateVf(pfNetdeviceName string) (string, error) {

	vfList, err := sriovnet.GetVfPciDevList(pfNetdeviceName)
	if err != nil {
		return "", err
	}
	for _, vf := range vfList {
		dirPath := filepath.Join("/sys/class/net", pfNetdeviceName, "device", vf, "net")
		fd, err := os.Open(dirPath)
		if err != nil {
			return "", err
		}
		defer fd.Close()
		fileInfos, err := fd.Readdir(-1)
		for i := range fileInfos {
			if fileInfos[i].Name() == "." || fileInfos[i].Name() == ".." {
				continue
			}
			vfNetdev := filepath.Join(dirPath, fileInfos[i].Name())
			_, err := os.Stat(vfNetdev)
			if err != nil {
				return "", err
			}
			return fileInfos[i].Name(), nil
		}
	}
	return "", fmt.Errorf("No VF are free")
}

func allocateVfForNetwork(userCmdArgs []string) (string, string, error) {

	networkName := getNonDefaultNetwork(userCmdArgs)

	network := getDockerNetworkResourceForName(networkName)
	if network == nil {
		return "", "", fmt.Errorf("Network not found")
	}

	pfNetdevName := network.Options["netdevice"]
	if pfNetdevName == "" {
		return "", "", fmt.Errorf("Netdevice invalid configuration")
	}

	vfNetdev, err := allocateVf(pfNetdevName)
	if err != nil {
		return "", "", err
	}

	rdmaDev, err := rdmamap.GetRdmaDeviceForNetdevice(vfNetdev)
	if err != nil {
		return "", "", err
	}
	return vfNetdev, rdmaDev, nil
}

func buildUserCmd(userCmdArgs []string) ([]string, error) {
	var runCmds []string
	var charDevCmdArgs []string

	runCmds = append(runCmds, "docker")
	runCmds = append(runCmds, "run")

	netDev, rdmaDev, err := allocateVfForNetwork(userCmdArgs)
	if err != nil {
		return nil, err
	}

	handle, err := netlink.LinkByName(netDev)
	if err != nil {
		return nil, err
	}
	netAttr := handle.Attrs()
	macAddr := netAttr.HardwareAddr.String()
	macAddrArg := "--mac-address=" + macAddr
	runCmds = append(runCmds, macAddrArg)

	charDevs := rdmamap.GetRdmaCharDevices(rdmaDev)
	if len(charDevs) != 0 {
		charDevCmdArgs = toCharDevCmdArgs(charDevs)
	}
	for _, devcmdArg := range charDevCmdArgs {
		runCmds = append(runCmds, devcmdArg)
	}

	runCmds = append(runCmds, "--cap-add=IPC_LOCK")

	for _, usrCmdArg := range userCmdArgs {
		runCmds = append(runCmds, usrCmdArg)
	}
	return runCmds, nil
}

func execUserRunCmd(userCmdArgs []string) {
	newCmd, err := buildUserCmd(userCmdArgs)
	if err != nil {
		fmt.Println("Fail to run docker container. Error= ", err)
		return
	}

	shellCmd := exec.Command("docker")
	shellCmd.Args = newCmd
	shellCmd.Stdout = os.Stdout
	shellCmd.Stdin = os.Stdin
	shellCmd.Stderr = os.Stderr
	shellCmd.Run()
}

func execRunCmd(cmd *cobra.Command, args []string) {
	if len(os.Args) <= 2 {
		cmd.HelpFunc()(cmd, os.Args)
		return
	}
	execUserRunCmd(os.Args[2:])
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Wrapper to docker run <command>",
	Run:   execRunCmd,
	// Ignore the errors for other command line arguments
	// that this program doesn't know about.
	// Refer https://github.com/spf13/cobra/pull/284
	// Refer https://github.com/spf13/cobra/pull/662/commits/96853a4e2c2716ef0059db31d147ab7e42a89d93#diff-2fc2009ba1969a36b69136d7fb7b2072R1690
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

var GitCommitId string

func versionCmdFunc(cmd *cobra.Command, args []string) {
	fmt.Println("Version:      ", appVersion)
	fmt.Println("Go version:   ", runtime.Version())
	fmt.Println("Git commit:   ", GitCommitId)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Wrapper to docker run <command>",
	Run:   versionCmdFunc,
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "docker_rdma_sriov [OPTIONS] COMMAND [ARG...]",
		Short:         "cli for managing docker rdma containers",
		SilenceUsage:  true,
		SilenceErrors: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.HelpFunc()(cmd, args)
			return nil
		},
	}
	return cmd
}

var level1Cmds = [...]*cobra.Command{
	versionCmd,
	runCmd,
	sriovCmds,
}

func setupCmds() *cobra.Command {
	rootCmd := newRootCmd()

	for _, cmds := range level1Cmds {
		rootCmd.AddCommand(cmds)
	}
	return rootCmd
}

func main() {
	rootCmd := setupCmds()
	rootCmd.Execute()
}

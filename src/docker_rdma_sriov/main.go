package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"runtime"
)

const (
	appVersion = "0.0.2"
)

var GitCommitId string

func getRightClientApiVersion() (string, error) {
	// Start with the lowest API to query which version is supported.
	lowestCli, err3 := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.12"))
	if err3 != nil {
		fmt.Println("Fail to create client: ", err3)
		return "", err3
	}
	allVersions, err2 := lowestCli.ServerVersion(context.Background())
	if err2 != nil {
		fmt.Println("Error to get server version: ", err2)
		return "", err2
	}
	return allVersions.APIVersion, nil
}

func getRightClient() (*client.Client, error) {
	var clientVersion string

	desiredVersion, err := getRightClientApiVersion()
	if err != nil {
		clientVersion = "unknown"
	} else {
		clientVersion = desiredVersion
	}
	cli, err2 := client.NewClientWithOpts(client.FromEnv, client.WithVersion(clientVersion))
	if err2 == nil {
		return cli, nil
	}
	return nil, err
}

func versionCmdFunc(cmd *cobra.Command, args []string) {
	var clientVersion string

	cli, err := getRightClient()
	if err == nil {
		clientVersion = cli.ClientVersion()
	}
	fmt.Println("Version:      ", appVersion)
	fmt.Println("Go version:   ", runtime.Version())
	fmt.Println("Git commit:   ", GitCommitId)
	fmt.Println("API version: ", clientVersion)
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
	statsCmd,
	devlinkCmds,
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

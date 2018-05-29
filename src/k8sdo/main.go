package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

const (
	appVersion = "0.0.1"
)

var GitCommitId string

func versionCmdFunc(cmd *cobra.Command, args []string) {
	fmt.Println("Version:      ", appVersion)
	fmt.Println("Go version:   ", runtime.Version())
	fmt.Println("Git commit:   ", GitCommitId)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Run:   versionCmdFunc,
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "k8sdo [OPTIONS] COMMAND [ARG...]",
		Short:         "cli for managing kubernetes cluster",
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
	installCmd,
	setupCmd,
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

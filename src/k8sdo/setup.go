package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var aipArg string
var modeArg string

func setupCmdFunc(cmd *cobra.Command, args []string) {
	fmt.Println("aip = ", aipArg)
	fmt.Println("mode = ", modeArg)

	if modeArg != "master" && modeArg != "node" {
		fmt.Println("Valid modes are master or node")
		return
	}
	err := kubeletCheckSriovCniTemplate()
	if err != nil {
		return
	}
	err = netfilterSetupIptables()
	if err != nil {
		return
	}
	err = kubeletConfigCgroupDriver()
	if err != nil {
		return
	}
	err = kubeletConfigDp(true)
	if err != nil {
		return
	}

	err = kubeadmInit(modeArg, aipArg)
	if err != nil {
		return
	}
	kubeletAllowMasterPodSchedule()
	kubectlSetupEnv()
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "setup kubernetes cluster",
	Run:   setupCmdFunc,
}

func init() {
	listFlags := setupCmd.Flags()
	listFlags.StringVarP(&aipArg, "aip", "a", "", "API server IP")
	listFlags.StringVarP(&modeArg, "mode", "m", "", "K8s mode (master or node)")
}

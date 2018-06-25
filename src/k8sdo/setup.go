package main

import (
	"log"
	"github.com/spf13/cobra"
)

var aipArg string
var modeArg string
var podSubnetArg string

func setupGenericNode() (error) {
	var err error

	err = netfilterSetupIptables()
	if err != nil {
		return err
	}
	err = kubeletConfigCgroupDriver()
	if err != nil {
		return err
	}
	err = kubeletConfigDp(true)
	return err
}

func setupMasterNode(aipArg string, podSubnetArg string) {
	var err error

	if aipArg == "" || podSubnetArg == "" {
		log.Println("Invalid API server IP or Pod subnet")
		return
	}

	err = setupGenericNode()
	if err != nil {
		return
	}

	err = kubeadmInit(aipArg, podSubnetArg)
	if err != nil {
		return
	}
	kubectlSetupEnv()
	kubectlGetNodes()
	kubeletAllowMasterPodSchedule()
	kubeletInstallSriovCni()
}

func setupWorkerNode() {
	err := setupGenericNode()
	if err != nil {
		return
	}
}

func setupCmdFunc(cmd *cobra.Command, args []string) {
	log.Println("Node mode = ", modeArg)
	log.Println("API Server IP = ", aipArg)
	log.Println("Pod CIDR network = ", podSubnetArg)

	if modeArg != "master" && modeArg != "node" {
		log.Println("Valid modes are master or node")
		return
	}
	if modeArg == "master" {
		setupMasterNode(aipArg, podSubnetArg)
	} else {
		setupWorkerNode()
	}
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "setup kubernetes cluster",
	Run:   setupCmdFunc,
}

func init() {
	listFlags := setupCmd.Flags()
	listFlags.StringVarP(&aipArg, "aip", "a", "", "K8s API server IP")
	listFlags.StringVarP(&modeArg, "mode", "m", "", "K8s mode (master or node)")
	listFlags.StringVarP(&podSubnetArg, "podnet", "p", "", "K8s Pod subnet (example: 193.168.1.0/24)")
}

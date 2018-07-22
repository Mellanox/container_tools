package main

import (
	"log"
	"github.com/spf13/cobra"
)

func startCmdFunc(cmd *cobra.Command, args []string) {
	log.Println("starting node")
	systemSwapoff()
	systemDisableSelinux()
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start kubernetes node",
	Run:   startCmdFunc,
}

func init() {

}

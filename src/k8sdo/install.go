package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"log"
)

func systemSwapoff() {
	log.Println("Turning swap off")
	execShellCmd("swapoff -a")
	log.Println("Turning swap off: done")
}

func systemDisableSelinux() {
	log.Println("Disabling selinux")
	execShellCmd("setenforce 0")
	log.Println("Disabling selinux: done")
}

func systemSetupKernelFile() {
	log.Println("Setting up kernel file")
	kernelVersion := strings.Trim(execShellCmdOutput("uname -r"), "\n")
	kernelCfgFile := filepath.Join("/boot", "config-"+kernelVersion)

	_, err := os.Stat(kernelCfgFile)
	if err != nil {
		srcFile := filepath.Join("/lib/modules", kernelVersion, "source/.config")
		fmt.Println("src file: ", srcFile)
		fmt.Println("dst file: ", kernelCfgFile)
		var cpcmd = []string {
			"cp", "-f", srcFile, kernelCfgFile,
		}
		execShellCmd(strings.Join(cpcmd, " "))
	}
	log.Println("Setting up kernel file: done")
}

var repofile = `[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-\$basearch
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg`

func k8sInstallRepoFile() {
	log.Println("Setting up K8s Repo file")
	_, err := os.Stat("/etc/yum.repos.d/kubernetes.repo")
	if err == nil {
		return
	}
	ioutil.WriteFile("/etc/yum.repos.d/kubernetes.repo", []byte(repofile), 0644)
	log.Println("Setting up K8s Repo file: done")
}

func k8sInstallPackages() {
	log.Println("Installing K8s packages")
	execShellCmd("yum install -y kubelet kubeadm kubectl kubernetes-cni")
	log.Println("Installing K8s packages: done")
}

func installCmdFunc(cmd *cobra.Command, args []string) {
	systemSwapoff()
	systemDisableSelinux()
	systemSetupKernelFile()
	k8sInstallRepoFile()
	k8sInstallPackages()
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install Kubernetes software",
	Run:   installCmdFunc,
}

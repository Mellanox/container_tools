package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	KUBELET_SRIOV_CNI_CONF_FILE = "/etc/cni/net.d/10-sriov-cni.conf"
)

func systemSwapoff() {
	execShellCmd("swapoff -a")
}

func systemDisableSelinux() {
	execShellCmd("setenforce 0")
}

func systemSetupKernelFile() {
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
}

func kubeletInstallSriovCni() {

	systemInstallGit()
	execShellCmd("mkdir -p /etc/cni/net.d")
	execShellCmd("mkdir -p /opt/cni/bin/")
	execShellCmd("git clone https://github.com/Mellanox/sriov-cni.git")
	srcFiles := filepath.Join("sriov-cni", "bin", getLinuxArch(), "*")
	fmt.Println("sriov cni src files: ", srcFiles)
	var cpcmd = []string {
		"cp", "-f", srcFiles, "/opt/cni/bin/",
	}
	execShellCmd(strings.Join(cpcmd, " "))
}

var repofile = `[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-\$basearch
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg`

func k8sInstallRepoFile() {
	_, err := os.Stat("/etc/yum.repos.d/kubernetes.repo")
	if err == nil {
		return
	}
	ioutil.WriteFile("/etc/yum.repos.d/kubernetes.repo", []byte(repofile), 0644)
}

func k8sInstallPackages() {
	execShellCmd("yum install -y kubelet-1.10.2-0.x86_64")
	execShellCmd("yum install -y kubectl-1.10.2-0.x86_64")
	execShellCmd("yum install -y kubeadm-1.10.2-0.x86_64")
	//execShellCmd("yum install -y kubelet kubeadm kubectl kubernetes-cni")
}

func systemInstallGit() {
	execShellCmd("yum install -y git")
}

var sriovCniTemplate =  `{
    "name": "mynet",
    "type": "sriov",
    "if0": "INVALID_IFACE",
    "ipam": {
        "type": "host-local",
        "subnet": "10.55.206.0/26",
        "routes": [
            { "dst": "0.0.0.0/0" }
        ],
        "gateway": "10.55.206.1"
    }
}`

func kubeletSetupSriovCniCfgTemplate() {
	ioutil.WriteFile(KUBELET_SRIOV_CNI_CONF_FILE, []byte(sriovCniTemplate), 0644)
	fmt.Printf("You must configure SRIOV CNI file %s\n", KUBELET_SRIOV_CNI_CONF_FILE)
}

func installCmdFunc(cmd *cobra.Command, args []string) {
	systemSwapoff()
	systemDisableSelinux()
	systemSetupKernelFile()
	k8sInstallRepoFile()
	k8sInstallPackages()
	kubeletInstallSriovCni()
	kubeletSetupSriovCniCfgTemplate()
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install kubernetes software",
	Run:   installCmdFunc,
}


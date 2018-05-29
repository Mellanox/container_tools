package main

import (
	"fmt"
	"strings"
	"os"
	"io/ioutil"
)

const (
	KUBELET_CFG_FILE = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
)

const (
	KUBEADM_INIT_WARNING    = "WARNING"
	KUBEADM_INIT_ERR        = "ERROR"
	KUBEADM_INIT_SUGGESTION = "Suggestion"
	KUBEADM_INIT_PREFLIGHT  = "[preflight]"
)

var ignoreWarnings = []string{
	"docker version is greater than",
	"kubelet service is not enabled",
	"crictl not found in system path",
}

var mustErrors = []string{
	"bridge-nf-call-iptables contents are not set to 1",
}

var kubeInitSuccessOutput = "Your Kubernetes master has initialized successfully"

func ignore_kubeadmInit_others(stderrLine string) bool {
	if strings.Contains(stderrLine, KUBEADM_INIT_PREFLIGHT) ||
		strings.Contains(stderrLine, KUBEADM_INIT_SUGGESTION) {
		return true
	} else {
		return false
	}
}

func check_kubeadmInit_errors(stderrLine string) bool {
	for _, errLine := range mustErrors {
		if strings.Contains(errLine, KUBEADM_INIT_ERR) == false {
			continue
		}
		if strings.Contains(errLine, errLine) == true {
			return true
		}
	}
	return false
}

func ignore_kubeadmInit_warnings(stderr string) bool {
	stderrLines := strings.Split(stderr, "\n")
	for _, line := range stderrLines {
		if ignore_kubeadmInit_others(line) == true {
			continue
		}
		if strings.Contains(line, KUBEADM_INIT_WARNING) == false {
			continue
		}
		for _, warning := range ignoreWarnings {
			if strings.Contains(line, warning) == false {
				return false
			}
		}
	}
	return true
}

func check_kubeadmInit_output(output string) error {

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, kubeInitSuccessOutput) {
			fmt.Println("Looks good")
			return nil
		}
	}
	return fmt.Errorf("Fail to perform kubelet init")
}

func kubeadmInit(mode string, aip string) error {
	var err error
	var apiServerIp = "--apiserver-advertise-address=" + aip
	var cmd []string

	cmd = append(cmd, "kubeadm")
	cmd = append(cmd, "init")
	cmd = append(cmd, apiServerIp)
	cmd = append(cmd, "--kubernetes-version")
	cmd = append(cmd, "stable-1.10")
	fmt.Println("Cmd = ", cmd)
	stdout, stderr, err1 := execUserCmd(cmd)
	if err1 != nil {
		fmt.Println("Command error = ", err1)
		return err1
	}
	fmt.Println("output = ", stdout)
	fmt.Println("err = ", stderr)

	if len(stderr) != 0 {
		ignore := ignore_kubeadmInit_warnings(stderr)
		if ignore == true {
			fmt.Println("err = ", stderr)
			return fmt.Errorf("Fail to do kubeadm init")
		}
		check := check_kubeadmInit_errors(stderr)
		if check == true {
			return fmt.Errorf("Error encountered")
		}
	}
	if len(stdout) != 0 {
		err = check_kubeadmInit_output(stdout)
	}
	return err
}

func kubeletGetCgroupConfig() (string, error) {
	line, err := readFileLineContains(KUBELET_CFG_FILE, "--cgroup-driver=")
	if err != nil {
		return "", err
	}
	driver, err := FindKeyValue(line, "driver")
	if err != nil {
		return "", err
	}
	fmt.Println("k8s cg driver = ", driver)
	return driver, nil
}

func updateK8sCgDriver(oldDriver, newDriver string) error {

	oldDriverCfg := "--cgroup-driver=" + oldDriver
	newDriverCfg := "--cgroup-driver=" + newDriver

	err := findReplaceFirstMatch(KUBELET_CFG_FILE, oldDriverCfg, newDriverCfg)
	return err
}

func kubeletConfigCgroupDriver() error {

	dockerCgDriver, err := dockerGetCgroupConfig()
	if err != nil {
		return err
	}
	fmt.Println("driver is:", dockerCgDriver)

	k8sCgDriver, err2 := kubeletGetCgroupConfig()
	if err2 != nil {
		return err
	}
	if dockerCgDriver == k8sCgDriver {
		return nil
	}
	err3 := updateK8sCgDriver(k8sCgDriver, dockerCgDriver)
	runSystemCtlReload()
	return err3
}

func kubeletUpdateDpFeatureGate(enable bool) {
	if enable == true {
		findReplaceFirstMatch(KUBELET_CFG_FILE, "DevicePlugins=false", "DevicePlugins=true")
	} else {
		findReplaceFirstMatch(KUBELET_CFG_FILE, "DevicePlugins=true", "DevicePlugins=false")
	}
}

func kubeletSetDpFeatureGate(setting string) {

	settingString := `Environment="KUBELET_EXTRA_ARGS=--feature-gates=DevicePlugins=` + setting + `"`
	appendToFileAtLine(KUBELET_CFG_FILE, settingString, 2)
}

func kubeletConfigDp(enable bool) error {
	_, err := readFileLineContains(KUBELET_CFG_FILE, "DevicePlugins=")
	if err == nil {
		//Past setting present, update it
		kubeletUpdateDpFeatureGate(enable)
	} else {
		//Fresh setting, so set it
		if enable == true {
			kubeletSetDpFeatureGate("true")
		} else {
			kubeletSetDpFeatureGate("false")
		}
	}
	runSystemCtlReload()
	return nil
}

func kubectlSetupEnv() {
	execShellCmd("mkdir -p $HOME/.kube")
	execShellCmd("cp -f /etc/kubernetes/admin.conf $HOME/.kube/config")
	execShellCmd("chown $(id -u):$(id -g) $HOME/.kube/config")
	appendToFile("/etc/profile", "export KUBECONFIG=/etc/kubernetes/kubelet.conf")
}

func kubeletCheckSriovCniTemplate() (error) {
	file := KUBELET_SRIOV_CNI_CONF_FILE
	_, err := os.Stat(file)
	if err != nil {
		fmt.Printf("sriov conf file in path %s doesn't exist.\n", KUBELET_SRIOV_CNI_CONF_FILE)
		fmt.Errorf("sriov conf file not present")
	}
	input, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		if strings.Contains(line, "INVALID_IFACE") {
			fmt.Printf("Please configure the sriov interface if0 in file %s\n",
				file)
			return fmt.Errorf("invalid sriov interface cfg")
		}
	}
	return nil
}

func kubeletAllowMasterPodSchedule() {
	execShellCmd("kubectl taint nodes --all node-role.kubernetes.io/master-")
	execShellCmd("systemctl restart kubelet")
}

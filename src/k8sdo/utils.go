package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"log"
)

// Takes command and argument as slice and returns stdout and stderr
func execUserCmd(userCmdArgs []string) (string, string, error) {

	log.Println("Executing:", strings.Join(userCmdArgs, " "))

	cmd := exec.Command(userCmdArgs[0])
	cmd.Args = userCmdArgs

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	cmd.Start()
	output, _ := ioutil.ReadAll(stdout)
	errout, _ := ioutil.ReadAll(stderr)
	cmd.Wait()

	return string(output), string(errout), nil
}

func execShellCmdInternal(cmd string, print bool) (string) {
	cmdArgs := strings.Split(cmd, " ")
	stdout, errout, _ := execUserCmd(cmdArgs)
	if stdout != "" && print {
		log.Println(stdout)
	}
	if errout != "" {
		log.Println("error is:", errout)
	}
	return stdout
}

func execShellCmd(cmd string) {
	execShellCmdInternal(cmd, true)
}

func execShellCmdOutput(cmd string) string {
	return execShellCmdInternal(cmd, false)
}

func writeStringToFile(file string, value string) error {

	fd, err := os.OpenFile(file, os.O_WRONLY, 0444)
	if err != nil {
		return err
	}
	defer fd.Close()
	fd.Seek(0, os.SEEK_SET)
	_, err = fd.WriteString(value)
	return err
}

func writeIntToFile(file string, value int) error {
	return writeStringToFile(file, strconv.Itoa(value))
}

func readFile(file string) (string, error) {
	fd, err := os.OpenFile(file, os.O_RDONLY, 0444)
	if err != nil {
		return "", err
	}
	defer fd.Close()
	fd.Seek(0, os.SEEK_SET)
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readFiletoLineSlices(file string) ([]string, error) {
	data, err := readFile(file)
	if err != nil {
		return nil, err
	}
	dataLines := strings.Split(data, "\n")
	return dataLines, nil
}

func readFileLineContains(file string, search string) (string, error) {
	data, err := readFiletoLineSlices(file)
	if err != nil {
		return "", err
	}
	for _, line := range data {
		if strings.Contains(line, search) == true {
			return line, nil
		}
	}
	return "", fmt.Errorf("search string not found")
}

func findReplaceFirstMatch(file string, old string, new string) error {
	stat, err := os.Stat(file)
	if err != nil {
		return err
	}

	input, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, old) {
			lines[i] = strings.Replace(line, old, new, 1)
			break
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(file, []byte(output), stat.Mode()&os.ModePerm)
	return err
}

func FindKeyValue(input string, key string) (string, error) {
	rex, err := regexp.Compile("(\\w+)=(\\w+)")
	if err != nil {
		return "", err
	}
	data := rex.FindAllStringSubmatch(input, -1)

	fmt.Println("input =", input)
	fmt.Println("data =", data)

	for _, kv := range data {
		k := kv[1]
		v := kv[2]
		if k == key {
			return v, nil
		}
	}
	return "", fmt.Errorf("Key %q not found", key)
}

func appendToFile(file string, entry string) {
	stat, err := os.Stat(file)
	if err != nil {
		return
	}

	input, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	lines := strings.Split(string(input), "\n")

	for _, line := range lines {
		if strings.Contains(line, entry) {
			fmt.Println("Entry exist in", file, line)
			return
		}
	}
	lines = append(lines, entry)
	lines = append(lines, "\n")
	output := strings.Join(lines, "\n")
	ioutil.WriteFile(file, []byte(output), stat.Mode()&os.ModePerm)
}

func appendToFileAtLine(file string, entry string, lineNumber int) {
	var newLines []string

	stat, err := os.Stat(file)
	if err != nil {
		return
	}

	input, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if (i + 1) == lineNumber {
			newLines = append(newLines, entry)
		}
		newLines = append(newLines, line)
	}
	output := strings.Join(newLines, "\n")
	ioutil.WriteFile(file, []byte(output), stat.Mode()&os.ModePerm)
}

func runSystemCtlReload() {
	var cmd = []string{"systemctl", "daemon-reload"}
	execUserCmd(cmd)
}

func netfilterSetupIptables() error {
	return writeIntToFile("/proc/sys/net/bridge/bridge-nf-call-iptables", 1)
}

func goarchToLinuxArch(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	case "ppc64le":
	default:
		return goarch
	}
	return goarch
}

func getLinuxArch() string {
	return goarchToLinuxArch(runtime.GOARCH)
}

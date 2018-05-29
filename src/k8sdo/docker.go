package main

import (
	"fmt"
	"strings"
)

func dockerGetCgroupConfig() (string, error) {
	var dockerInfo = []string{"docker", "info"}

	stdout, _, err := execUserCmd(dockerInfo)
	if err != nil {
		return "", err
	}
	stdoutLines := strings.Split(stdout, "\n")
	for _, line := range stdoutLines {
		if strings.Contains(line, "Cgroup Driver: ") != true {
			continue
		}
		cgInfo := strings.Split(line, ":")
		if len(cgInfo) != 2 {
			return "", fmt.Errorf("Error parsing cgroup info")
		}
		cgdriver := strings.Trim(cgInfo[1], " ")
		if cgdriver == "cgroupfs" || cgdriver == "systemd" {
			return cgdriver, nil
		}
	}
	return "", fmt.Errorf("Fail to find key")
}

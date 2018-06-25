package main

import (
	"fmt"
	"github.com/Mellanox/sriovnet"
	"os"
	"path/filepath"
	"strings"
)

const (
	NetdevPhysSwitchIdFile = "phys_switch_id"
)

const (
	udevRulesFile = "/etc/udev/rules.d/82-net-setup-link.rules"
)

func GetNetdevPhysSwitchId(netdev string) (string, error) {

	file := filepath.Join(sriovnet.NetSysDir, netdev, NetdevPhysSwitchIdFile)
	fileObj := fileObject{
		Path: file,
	}

	data, err := fileObj.Read()
	if err != nil {
		return "", err
	} else {
		return strings.Trim(data, "\n"), nil
	}
}

func delNetdevUdevEntry(netdev string) {
	var outlines []string

	lines, err := ReadFileToLines(udevRulesFile)
	if err != nil {
		return
	}
	for _, line := range lines {
		fmt.Println(line)
		if !strings.Contains(line, netdev) {
			outlines = append(outlines, line)
		}
	}
	output := strings.Join(outlines, "\n")
	os.Remove(udevRulesFile)
	AppendStringToFile(udevRulesFile, output)
}

func addNetdevUdevEntry(netdev string, id string) {
	s := fmt.Sprintf(`SUBSYSTEM=="net", ACTION=="add", ATTR{phys_switch_id}=="%s", ATTR{phys_port_name}!="", NAME="%s_$attr{phys_port_name}"`, id, netdev)

	s = s + "\n"

	AppendStringToFile(udevRulesFile, s)
}

func setupSwitchdevUdevRule(netdev string) error {
	id, err := GetNetdevPhysSwitchId(netdev)
	if err != nil {
		return err
	}
	_, err1 := os.Stat(udevRulesFile)
	if err1 == nil {
		fileObj := fileObject{
			Path: udevRulesFile,
		}

		data, err := fileObj.Read()
		if err != nil {
			return err
		}
		if strings.Contains(data, netdev) {
			delNetdevUdevEntry(netdev)
		}
	}
	addNetdevUdevEntry(netdev, id)
	return nil
}

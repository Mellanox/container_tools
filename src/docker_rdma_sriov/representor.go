package main

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"strconv"
)

func SetVfRepresentorLinkUp(pfNetdevName string, vfIndex int) error {

	vfRepNetdevName := pfNetdevName + "_" + strconv.Itoa(vfIndex)
	fmt.Println("Vf rep:", vfRepNetdevName)

	handle, err := netlink.LinkByName(vfRepNetdevName)
	if err != nil {
		return err
	}
	return netlink.LinkSetUp(handle)
}

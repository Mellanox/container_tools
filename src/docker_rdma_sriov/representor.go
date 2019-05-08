package main

import (
	"fmt"
	"github.com/Mellanox/sriovnet"
	"github.com/vishvananda/netlink"
	"strconv"
)

func getVfRepresentorLink(pfNetdevName string, vfIndex int) (netlink.Link, error) {

	/* First try to get old way eth0_vfIndex */
	vfRepNetdevName := pfNetdevName + "_" + strconv.Itoa(vfIndex)
	handle, err := netlink.LinkByName(vfRepNetdevName)
	if err == nil {
		fmt.Println("Vf rep:", vfRepNetdevName)
		return handle, nil
	}

	/* Possibly new OS with eth0_pfAvfN way scheme */
	vfRepNetdevName, err = sriovnet.GetVfRepresentor(pfNetdevName, vfIndex)
	if err != nil {
		fmt.Println("Representor not found netdev for pf = %v vf %d\n",
			pfNetdevName, vfIndex)
		return nil, err
	}
	fmt.Println("Vf rep:", vfRepNetdevName)
	return netlink.LinkByName(vfRepNetdevName)
}

func SetVfRepresentorLinkUp(pfNetdevName string, vfIndex int) error {

	handle, err := getVfRepresentorLink(pfNetdevName, vfIndex)
	if err != nil {
		return err
	}
	return netlink.LinkSetUp(handle)
}

package main

import (
	"fmt"

	"github.com/drtimf/wmi"
)

func main() {
	var service *wmi.Service
	var err error

	if service, err = wmi.NewLocalService(wmi.RootCIMV2); err != nil {
		panic(err)
	}

	defer service.Close()

	var enum *wmi.Enum
	if enum, err = service.ExecQuery(`SELECT InterfaceIndex, Manufacturer, MACAddress, Name FROM Win32_NetworkAdapter`); err != nil {
		panic(err)
	}

	defer enum.Close()

	for {
		var instance *wmi.Instance
		if instance, err = enum.Next(); err != nil {
			panic(err)
		}

		if instance == nil {
			break
		}

		defer instance.Close()

		var val interface{}
		var interfaceIndex int32
		var manufacturer, MACAddress, name string

		if val, _, _, err = instance.Get("InterfaceIndex"); err != nil {
			panic(err)
		}

		if val != nil {
			interfaceIndex = val.(int32)
		}

		if val, _, _, err = instance.Get("Manufacturer"); err != nil {
			panic(err)
		}

		if val != nil {
			manufacturer = val.(string)
		}

		if val, _, _, err = instance.Get("MACAddress"); err != nil {
			panic(err)
		}

		if val != nil {
			MACAddress = val.(string)
		}

		if val, _, _, err = instance.Get("Name"); err != nil {
			panic(err)
		}

		if val != nil {
			name = val.(string)
		}

		fmt.Printf("%6d %-25s%20s\t%s\n", interfaceIndex, manufacturer, MACAddress, name)
	}
}

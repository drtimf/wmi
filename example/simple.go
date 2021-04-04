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

	var computerSystem wmi.Win32_ComputerSystem
	if err = service.Query("SELECT * FROM Win32_ComputerSystem", &computerSystem); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", computerSystem)
}

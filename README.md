# wmi
A wrapper for local and remote Windows WMI at both low level calls to COM, and at a high level Go object mapping.

There are a number of WMI library implementations around, but not many of them provide:
* Both local and remote WMI access to the WMI provider
* A single session to execute many queries
* Low level access to the WMI API
* High level mapping of WMI objects to Go objects
* WMI method execution

This presently only works on Windows.  If there is ever a port of the Python Impacket to Go, it would be good to have
this work on Linux and MacOS as well.

## A Simple Example

```go
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
```



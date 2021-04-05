# wmi
A wrapper for local and remote Windows WMI at both low level calls to COM, and at a high level Go object mapping.

There are a number of WMI library implementations around, but not many of them provide:
* Both local and remote access to the WMI provider
* A single session to execute many queries
* Low level access to the WMI API
* High level mapping of WMI objects to Go objects
* WMI method execution

This presently only works on Windows.  If there is ever a port of the Python Impacket to Go, it would be good to have this work on Linux and MacOS as well.

[![Build Status](https://travis-ci.com/drtimf/wmi.svg?branch=main)](https://travis-ci.org/drtimf/wmi)


## Examples

### A Simple High-Level Example to Query Win32_ComputerSystem

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

### A Low-Level Example to Query Win32_NetworkAdapter

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
```

## Usage

Objects in this API include a Close() method which should be called when the object is no longer required.  This is important as it invokes the COM Release() to free the resources memory.

There are three high level objects:
* Service: A connection to either a local or remote WMI service
* Enum: An enumerator of WMI instances
* Instance: An instance of a WMI class

### Open a Connection to a WMI Service and Namespace

In each case a new *wmi.Service is created which can be used to obtain class instances and execute queries.

Open the local WMI provider:
``` go
func NewLocalService(namespace string) (s *Service, err error)
````

Open a connection to a remote WMI service with a username and a password:
``` go
func NewRemoteService(server string, namespace string, username string, password string) (s *Service, err error)
```

Open a child namespace of an existing service:
``` go
func (s *Service) OpenNamespace(namespace string) (newService *Service, err error)
```

### Get a Single WMI Object

Obtain a single WMI object given its path:
``` go
func (s *Service) GetObject(objectPath string) (instance *Instance, err error)
```

### Query WMI Objects

Obtain a WMI enumerator for a class of a given name:
``` go
func (s *Service) CreateInstanceEnum(className string) (e *Enum, err error)
```

Enumerate a WMI class of a given name and map the objects to a structure or slice of structures:
``` go
func (s *Service) ClassInstances(className string, dst interface{}) (err error)
```

Execute a WMI Query Language (WQL) query and return an enumerator for the queried class instances:
``` go
func (s *Service) ExecQuery(wqlQuery string) (e *Enum, err error)
```

Execute a WMI Query Language (WQL) query and map the results to a structure or slice of structures:
``` go
func (s *Service) Query(query string, dst interface{}) (err error)
```

### Execute a WMI Method

``` go
func (s *Service) ExecMethod(className string, methodName string, inParams *Instance) (outParam *Instance, err error)
```

### Obtain the Next Object from a WMI Enumerator

``` go
func (e *Enum) Next() (instance *Instance, err error)
```

``` go
func (e *Enum) NextObject(dst interface{}) (done bool, err error)
```


### ... Instance
``` go
func (i *Instance) GetClassName() (className string, err error)
```


``` go
func (i *Instance) SpawnInstance() (instance *Instance, err error)
```


``` go
func (i *Instance) GetNames() (names []string, err error)
```


``` go
func (i *Instance) Get(name string) (value interface{}, cimType CIMTYPE_ENUMERATION, flavor WBEM_FLAVOR_TYPE, err error)
```

``` go
func (i *Instance) GetPropertyValue(name string) (value string, err error)
```



``` go
func (i *Instance) Put(name string, value interface{}) (err error)
```


``` go
func (i *Instance) BeginEnumeration() (err error)
```

``` go
func (i *Instance) NextAsVariant() (done bool, name string, value *ole.VARIANT, cimType CIMTYPE_ENUMERATION, flavor WBEM_FLAVOR_TYPE, err error)
```

``` go
func (i *Instance) Next() (done bool, name string, value interface{}, cimType CIMTYPE_ENUMERATION, flavor WBEM_FLAVOR_TYPE, err error)
```

``` go
func (i *Instance) EndEnumeration() (err error)
```

``` go
func (i *Instance) GetProperties() (properties []Property, err error)
```



``` go
func (i *Instance) BeginMethodEnumeration() (err error)
```

``` go
func (i *Instance) NextMethod() (done bool, name string, err error)
```

``` go
func (i *Instance) EndMethodEnumeration() (err error)
```

``` go
func (i *Instance) GetMethods() (methodNames []string, err error)
```



``` go
func (i *Instance) GetMethod(methodName string) (inSignature *Instance, outSignature *Instance, err error)
```

``` go
func (i *Instance) GetMethodParameters(methodName string) (inParam *Instance, err error)
```





// +build windows
package wmi

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
)

// Currently dumps a lot of stuff to the console...
// REVISIT: Need memory leak tests!!

func evaluateClassEnum(t *testing.T, enum *Enum) {
	var err error

	for {
		var instance *Instance
		instance, err = enum.Next()
		assert.NilError(t, err, "Failed to get next object from an enumeration")

		if instance == nil {
			break
		}
		defer instance.Close()

		evaluateInstance(t, instance)
	}
}

func evaluateInstance(t *testing.T, instance *Instance) {
	var err error

	var className string
	className, err = instance.GetClassName()
	assert.NilError(t, err, "Failed to get the objects class name")
	assert.Assert(t, className != "", "Class name is empty")
	fmt.Println(className)

	var properties []Property
	properties, err = instance.GetProperties()
	assert.NilError(t, err, "Failed to get object properties")
	for _, p := range properties {
		if p.Value != nil {
			assert.Assert(t, p.Name != "", "Property name is empty")
			fmt.Printf("\t%s = %s\n", p.Name, p.ValueAsString())
		} else {
			fmt.Printf("\t%s = <nil>\n", p.Name)
		}
	}
}

func TestLocalWMIClassQueries(t *testing.T) {
	var err error

	// Open the local root WMI namespace
	var rootService, service *Service
	rootService, err = NewLocalService(`ROOT`)
	assert.NilError(t, err, "Failed to open root namespace")
	defer rootService.Close()

	// Namespace enum
	var namespaceEnum *Enum
	namespaceEnum, err = rootService.CreateInstanceEnum("__NAMESPACE")
	assert.NilError(t, err, "Failed to begin enumeration of namespaces")
	defer namespaceEnum.Close()

	evaluateClassEnum(t, namespaceEnum)

	// Open the cimv2 child namespace
	service, err = rootService.OpenNamespace(`cimv2`)
	assert.NilError(t, err, "Failed to open child namespace")
	defer service.Close()

	// Instance enum
	var enum *Enum
	enum, err = service.CreateInstanceEnum("Win32_OperatingSystem")
	assert.NilError(t, err, "Failed to begin enumeration of object")
	defer enum.Close()

	evaluateClassEnum(t, enum)

	// Query enum
	enum, err = service.ExecQuery("SELECT Name, Caption, Description, Domain, Manufacturer, Model, NumberOfProcessors, NumberOfLogicalProcessors FROM Win32_ComputerSystem")
	assert.NilError(t, err, "Failed to execute query")
	defer enum.Close()

	evaluateClassEnum(t, enum)
}

func TestLocalWMIMethods(t *testing.T) {
	var err error

	// Open cimv2 WMI
	var service *Service
	service, err = NewLocalService(RootCIMV2)
	assert.NilError(t, err, "Failed to open service")
	defer service.Close()

	// Get the WMI registry provider
	var instance *Instance
	instance, err = service.GetObject("StdRegProv")
	assert.NilError(t, err, "Failed to get object")
	defer instance.Close()

	// List and check registry methods
	var methodNames []string
	methodNames, err = instance.GetMethods()
	assert.NilError(t, err, "Failed to get methods")
	fmt.Println(methodNames)
	assert.Assert(t, len(methodNames) > 0, "No method names were returned")

	var found int
	for _, m := range methodNames {
		if m == "CreateKey" || m == "EnumKey" || m == "EnumValues" || m == "GetStringValue" {
			found++
		}
	}
	assert.Equal(t, found, 4, "Did not find all named methods")

	// Call WMI registry EnumValues
	var inSignature *Instance
	var outSignature *Instance
	inSignature, outSignature, err = instance.GetMethod("EnumValues")
	assert.NilError(t, err, "Failed to get the EnumValues method")
	defer inSignature.Close()
	defer outSignature.Close()

	evaluateInstance(t, inSignature)
	evaluateInstance(t, outSignature)

	var inParam *Instance
	inParam, err = inSignature.SpawnInstance() // C examples online have this, but it does not appear to be needed...
	assert.NilError(t, err, "Failed to spawn a new instance")
	defer inParam.Close()

	evaluateInstance(t, inParam)

	err = inParam.Put("hDefKey", int(HKEY_LOCAL_MACHINE))
	assert.NilError(t, err, "Failed to put the hDefKey")
	err = inParam.Put("sSubKeyName", `SYSTEM\CurrentControlSet\Control`)
	assert.NilError(t, err, "Failed to put the sSubKeyName")

	var outParam *Instance
	outParam, err = service.ExecMethod("StdRegProv", "EnumValues", inParam)
	assert.NilError(t, err, "Failed to execute the EnumValues method")
	defer outParam.Close()

	evaluateInstance(t, outParam)
}

func TestLocalMethodExecutor(t *testing.T) {
	var err error

	// Open cimv2 WMI
	var service *Service
	service, err = NewLocalService(RootCIMV2)
	assert.NilError(t, err, "Failed to open service")
	defer service.Close()

	// Get Win32_Process object
	var processInstance *Instance
	processInstance, err = service.GetObject("Win32_Process")
	assert.NilError(t, err, "Failed to get Win32_Process object")
	defer processInstance.Close()

	// List Win32_Process methods
	fmt.Printf("Win32_Process\n")
	var methodNames []string
	methodNames, err = processInstance.GetMethods()
	assert.NilError(t, err, "Failed to get methods")
	fmt.Printf("\t%v\n", methodNames)
	assert.Assert(t, len(methodNames) > 0, "No method names were returned")

	// Get a Win32_ProcessStartup object and span an instance
	var processStartupInstanceClass *Instance
	processStartupInstanceClass, err = service.GetObject("Win32_ProcessStartup")
	assert.NilError(t, err, "Failed to get Win32_ProcessStartup object")
	defer processStartupInstanceClass.Close()

	var processStartupInstance *Instance
	processStartupInstance, err = processStartupInstanceClass.SpawnInstance()
	assert.NilError(t, err, "Failed to spawn a Win32_ProcessStartup instance")
	processStartupInstance.Put("ShowWindow", 0)
	assert.NilError(t, err, "Failed to set ShowWindow on Win32_ProcessStartup object")

	evaluateInstance(t, processStartupInstance)

	// Launch a process
	fmt.Println("Win32_Process Create()")
	var pid interface{}
	err = BeginMethodExecute(service, processInstance, "Win32_Process", "Create").
		Set("CommandLine", "netstat.exe -anop tcp").Set("CurrentDirectory", `C:\`).Set("ProcessStartupInformation", processStartupInstance).
		Execute().
		Get("ProcessId", &pid).
		End()
	assert.NilError(t, err, "Failed to execite command")
	assert.Assert(t, pid.(int32) != 0)
	fmt.Println("\tProcessId:", pid)

	// Check launched process
	var processEnum *Enum
	processEnum, err = service.ExecQuery(fmt.Sprintf("SELECT Name, Caption, CommandLine, CreationDate, Description, ExecutablePath, ProcessId FROM Win32_Process WHERE ProcessId = %v", pid))
	assert.NilError(t, err, "Failed to get process")
	evaluateClassEnum(t, processEnum)
}

func TestLocalWMIRegistry(t *testing.T) {
	var err error

	// Open cimv2 WMI
	var service *Service
	service, err = NewLocalService(RootCIMV2)
	assert.NilError(t, err, "Failed to open service")
	defer service.Close()

	// Open WMI registry
	var wreg *Registry
	wreg, err = NewRegistry(service)
	assert.NilError(t, err, "Failed to open WMI registry")
	defer wreg.Close()

	// Get values from a registry path
	var vals []RegistryValue
	vals, err = wreg.EnumValues(HKEY_LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control`)
	assert.NilError(t, err, "Failed to enum SYSTEM\\CurrentControlSet\\Control registry values")

	// Enumerate value list and get each value
	fmt.Println(`SYSTEM\CurrentControlSet\Control`)
	for _, v := range vals {
		var val interface{}
		val, err = wreg.GetValue(HKEY_LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control`, v.Type, v.Name)
		assert.NilError(t, err, "Failed to get a registry value")
		fmt.Printf("\t%s (%s) = %v\n", v.Name, RegTypeToString(v.Type), val)
	}

	vals, err = wreg.EnumValues(HKEY_LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\IntegrityServices`)
	assert.NilError(t, err, "Failed to enum SYSTEM\\CurrentControlSet\\Control\\IntegrityServices registry values")

	// Enumerate value list and get each value
	fmt.Println(`SYSTEM\CurrentControlSet\Control\IntegrityServices`)
	for _, v := range vals {
		var val interface{}
		val, err = wreg.GetValue(HKEY_LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\IntegrityServices`, v.Type, v.Name)
		assert.NilError(t, err, "Failed to get a registry value")
		switch xv := val.(type) {
		case []uint8:
			fmt.Printf("\t%s (%s) = %v....\n", v.Name, RegTypeToString(v.Type), xv[:20])
		default:
			fmt.Printf("\t%s (%s) = %v\n", v.Name, RegTypeToString(v.Type), val)
		}
	}
}

func TestLocalObjectMappingQuery(t *testing.T) {

	var err error

	// Open cimv2 WMI
	var service *Service
	service, err = NewLocalService(RootCIMV2)
	assert.NilError(t, err, "Failed to open service")
	defer service.Close()

	fmt.Println("Win32_ComputerSystem")
	var computerSystem Win32_ComputerSystem
	err = service.Query("SELECT * FROM Win32_ComputerSystem", &computerSystem)
	assert.NilError(t, err, "Failed to query Win32_ComputerSystem")
	fmt.Printf("%+v\n", computerSystem)

	fmt.Println("Directory of C:")
	var files []CIM_DataFile
	err = service.Query(`SELECT * FROM CIM_DataFile WHERE Drive = 'C:' AND Path = '\\'`, &files)
	assert.NilError(t, err, "Failed to query CIM_DataFile")
	assert.Assert(t, len(files) > 0, "No files were returned")
	for _, f := range files {
		fmt.Printf("%16d %s %-8v %s\n", f.FileSize, f.LastModified, f.Hidden, f.FileName)
	}

	fmt.Println("List of Win32_NetworkAdapter")
	var enum *Enum
	enum, err = service.CreateInstanceEnum("Win32_NetworkAdapter")
	assert.NilError(t, err, "Failed to create an enumerator on Win32_NetworkAdapter objects")
	defer enum.Close()

	for done := false; !done; {
		var networkAdapter Win32_NetworkAdapter
		done, err = enum.NextObject(&networkAdapter)
		assert.NilError(t, err, "Failed to get next Win32_NetworkAdapter")

		if !done {
			fmt.Printf("%6d %-25s%20s\t%s\n", networkAdapter.InterfaceIndex, networkAdapter.Manufacturer, networkAdapter.MACAddress, networkAdapter.Name)
		}
	}
}

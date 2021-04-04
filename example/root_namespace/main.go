package main

import (
	"fmt"

	"github.com/drtimf/wmi"
)

func main() {
	var rootService *wmi.Service
	var err error

	if rootService, err = wmi.NewLocalService("ROOT"); err != nil {
		panic(err)
	}

	defer rootService.Close()

	type Namespace struct {
		Name string
	}

	var namespaces []Namespace
	if err = rootService.ClassInstances("__NAMESPACE", &namespaces); err != nil {
		panic(err)
	}

	fmt.Println("ROOT Namespaces:")
	for _, n := range namespaces {
		fmt.Printf("\t%s\n", n.Name)
	}
}

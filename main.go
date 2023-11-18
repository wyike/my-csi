package main

import (
	"flag"
	"fmt"

	"github.com/wyike/my-csi/pkg/driver"
)

func main() {
	var (
		endpoint     = flag.String("endpoint", "default", "Endpoint our gPRC server would run at")
		vchost       = flag.String("vc-host", "10.8.8.8", "vcenter ipaddress")
		vcuser       = flag.String("vc-user", "acutie", "vcenter access user")
		vcpassword   = flag.String("vc-password", "nihao", "vcenter access password")
		vcdatacenter = flag.String("vc-datacenter", "dc0", "vcenter datacenter to use")
		vcdatastore  = flag.String("vc-datastore", "vsanDatastore", "vcenter datastore to use available for your hosts")
		nodeplugin   = flag.Bool("node-plugin", false, "serve node plugin functionalities or not")
	)

	flag.Parse()
	fmt.Println(*endpoint, *vchost, *vcdatacenter, *vcdatastore)

	// create a driver instance
	drv, err := driver.NewDriver(&driver.InputParams{
		Name:         driver.DefaultName,
		Endpoint:     *endpoint,
		VcHost:       *vchost,
		VcUser:       *vcuser,
		VcPassword:   *vcpassword,
		VcDatacenter: *vcdatacenter,
		VcDatastore:  *vcdatastore,
		NodePlugin:   *nodeplugin,
	})
	if err != nil {
		panic(fmt.Sprintf("fail to initialize the driver: %s", err.Error()))
	}

	// run on that driver instance, it would start the gRPC server
	if err := drv.Run(); err != nil {
		panic(fmt.Sprintf("fail to run the driver: %s", err.Error()))
	}
}

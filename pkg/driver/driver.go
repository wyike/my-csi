package driver

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/wyike/my-csi/pkg/vsphere"
)

const (
	DefaultName = "my-csi"
)

type Driver struct {
	name     string
	region   string
	endpoint string

	srv *grpc.Server

	ready bool

	vsphereClient *vsphere.VsphereClient
}

type InputParams struct {
	Name     string
	Endpoint string

	VcHost       string
	VcUser       string
	VcPassword   string
	VcDatacenter string
	VcDatastore  string

	NodePlugin bool
}

// NewDriver initialize a csi driver
func NewDriver(params *InputParams) (*Driver, error) {
	var vsphereClient *vsphere.VsphereClient = nil

	if !params.NodePlugin {
		fmt.Printf("Initialize vsphere client with credentials\n")
		vc, err := vsphere.NewVsphereClient(params.VcHost, params.VcUser, params.VcPassword, params.VcDatacenter, params.VcDatastore)
		if err != nil {
			return nil, fmt.Errorf("fail to create vsphere client, %s", err.Error())
		}
		vsphereClient = vc
	}

	return &Driver{
		name:          params.Name,
		endpoint:      params.Endpoint,
		vsphereClient: vsphereClient,
	}, nil
}

// Run starts the grpc server
func (d *Driver) Run() error {
	url, err := url.Parse(d.endpoint)
	if err != nil {
		return fmt.Errorf("fail to parse the endpoint: %s", err.Error())
	}

	if url.Scheme != "unix" {
		return fmt.Errorf("unsupported scheme %s, only supported scheme is unix", url.Scheme)
	}

	grpcAddress := path.Join(url.Host, filepath.FromSlash(url.Path))
	if url.Host == "" {
		grpcAddress = filepath.FromSlash(url.Path)
	}

	if err := os.Remove(grpcAddress); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("fail to remove the listen address, : %s", err.Error())
	}

	listener, err := net.Listen(url.Scheme, grpcAddress)
	if err != nil {
		return fmt.Errorf("fail to initialized listener: %s", err.Error())
	}

	d.srv = grpc.NewServer()
	csi.RegisterControllerServer(d.srv, d)
	csi.RegisterNodeServer(d.srv, d)
	csi.RegisterIdentityServer(d.srv, d)

	d.ready = true

	// start grpc server
	return d.srv.Serve(listener)
}

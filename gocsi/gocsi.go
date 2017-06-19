package gocsi

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"plugin"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/container-storage-interface/libraries/gocsi/csi"
)

var (
	initOnce      sync.Once
	endpointCtors = map[string]func() interface{}{}
)

// Init initializes the CSI endpoint manager.
func Init(ctx context.Context) error {
	var err error
	initOnce.Do(func() {
		err = loadSharedObjects(ctx)
	})
	return err
}

// Endpoint is a gRPC server that provides the CSI Controller,
// Identity, and Node services.
type Endpoint interface {
	Init(ctx context.Context) error
	Serve(ctx context.Context, li net.Listener) error
	Shutdown(ctx context.Context) error
}

// Service is one configuration of a CSI endpoint's services.
type Service interface {
	Endpoint
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
}

type endpoint struct {
	once sync.Once
	name string
	endp Endpoint
	conn *pipeConn
	clnt *grpc.ClientConn
}

var errInvalidEndpointProvider = fmt.Errorf("invalid endpoint provider")

// New returns a CSI endpoint for the specified provider. If no
// provider matches the specified name a nil value is returned.
func New(ctx context.Context, name string) (Service, error) {

	// ensure the package is initialized and the shared objects
	// are loaded and available
	if err := Init(ctx); err != nil {
		return nil, err
	}

	for k, v := range endpointCtors {
		if strings.EqualFold(k, name) {
			o := v()
			if e, ok := o.(Endpoint); ok {
				return &endpoint{
					name: k,
					endp: e,
					conn: newPipeConn(k),
				}, nil
			}
			return nil, fmt.Errorf("invalid endpoint type: %T", o)
		}
	}

	return nil, errInvalidEndpointProvider
}

func (e *endpoint) Init(ctx context.Context) error {
	return e.endp.Init(ctx)
}

// Serve starts the piped connection to the Go plug-in that provides
// the implementation of the CSI services.
func (e *endpoint) Serve(
	ctx context.Context, li net.Listener) (err error) {

	return e.endp.Serve(ctx, e.conn)
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections.
func (e *endpoint) Shutdown(ctx context.Context) error {
	e.endp.Shutdown(ctx)
	e.conn.Close()
	return nil
}

func (e *endpoint) dial(
	ctx context.Context) (client *grpc.ClientConn, err error) {

	return grpc.DialContext(
		ctx,
		e.name,
		grpc.WithInsecure(),
		grpc.WithDialer(e.conn.Dial))
}

func (e *endpoint) dialController(
	ctx context.Context) (csi.ControllerClient, error) {

	c, err := e.dial(ctx)
	if err != nil {
		return nil, err
	}
	return csi.NewControllerClient(c), nil
}

func (e *endpoint) dialIdentity(
	ctx context.Context) (csi.IdentityClient, error) {

	c, err := e.dial(ctx)
	if err != nil {
		return nil, err
	}
	return csi.NewIdentityClient(c), nil
}

func (e *endpoint) dialNode(
	ctx context.Context) (csi.NodeClient, error) {

	c, err := e.dial(ctx)
	if err != nil {
		return nil, err
	}
	return csi.NewNodeClient(c), nil
}

////////////////////////////////////////////////////////////////////////////////
//                               Go Plug-ins                                  //
////////////////////////////////////////////////////////////////////////////////

func loadSharedObjects(ctx context.Context) error {
	// read the paths of the go plug-in files
	rdr := csv.NewReader(strings.NewReader(os.Getenv("CSI_PLUGINS")))
	sos, err := rdr.Read()
	if err != nil && err != io.EOF {
		return err
	}
	if len(sos) == 0 {
		return nil
	}

	// iterate the shared object files and load them one at a time
	for _, so := range sos {

		// attempt to open the plug-in
		p, err := plugin.Open(so)
		if err != nil {
			return err
		}
		log.Printf("loaded plug-in: %s\n", so)

		epsSym, err := p.Lookup("Endpoints")
		if err != nil {
			return err
		}
		eps, ok := epsSym.(*map[string]func() interface{})
		if !ok {
			return fmt.Errorf("error: invalid endpoints field: %T", epsSym)
		}

		if eps == nil {
			return fmt.Errorf("error: nil endpoints")
		}

		// record the endpoint provider names and constructors
		for k, v := range *eps {
			endpointCtors[k] = v
			log.Printf("registered endpoint: %s\n", k)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                                PipeConn                                    //
////////////////////////////////////////////////////////////////////////////////

func newPipeConn(name string) *pipeConn {
	return &pipeConn{
		addr: &pipeAddr{name: name},
		chcn: make(chan net.Conn),
	}
}

type pipeConn struct {
	addr *pipeAddr
	chcn chan net.Conn
}

func (p *pipeConn) Dial(
	raddr string,
	timeout time.Duration) (net.Conn, error) {

	r, w := net.Pipe()
	go func() {
		p.chcn <- r
	}()

	return w, nil
}

func (p *pipeConn) Accept() (net.Conn, error) {
	for c := range p.chcn {
		log.Printf("%s.Accept\n", p.addr.name)
		return c, nil
	}
	return nil, nil
}

func (p *pipeConn) Close() error {
	close(p.chcn)
	return nil
}

func (p *pipeConn) Addr() net.Addr {
	return p.addr
}

type pipeAddr struct {
	name string
}

func (a *pipeAddr) Network() string {
	return "modcsi"
}

func (a *pipeAddr) String() string {
	return a.name
}

////////////////////////////////////////////////////////////////////////////////
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

func (e *endpoint) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	c, err := e.dialController(ctx)
	if err != nil {
		return nil, err
	}
	return c.CreateVolume(ctx, req)
}

func (e *endpoint) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	c, err := e.dialController(ctx)
	if err != nil {
		return nil, err
	}
	return c.DeleteVolume(ctx, req)
}

func (e *endpoint) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	c, err := e.dialController(ctx)
	if err != nil {
		return nil, err
	}
	return c.ControllerPublishVolume(ctx, req)
}

func (e *endpoint) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	c, err := e.dialController(ctx)
	if err != nil {
		return nil, err
	}
	return c.ControllerUnpublishVolume(ctx, req)
}

func (e *endpoint) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	c, err := e.dialController(ctx)
	if err != nil {
		return nil, err
	}
	return c.ValidateVolumeCapabilities(ctx, req)
}

func (e *endpoint) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	c, err := e.dialController(ctx)
	if err != nil {
		return nil, err
	}
	return c.ListVolumes(ctx, req)
}

func (e *endpoint) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	c, err := e.dialController(ctx)
	if err != nil {
		return nil, err
	}
	return c.GetCapacity(ctx, req)
}

func (e *endpoint) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	c, err := e.dialController(ctx)
	if err != nil {
		return nil, err
	}
	return c.ControllerGetCapabilities(ctx, req)
}

////////////////////////////////////////////////////////////////////////////////
//                             Identity Service                               //
////////////////////////////////////////////////////////////////////////////////

func (e *endpoint) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	c, err := e.dialIdentity(ctx)
	if err != nil {
		return nil, err
	}
	return c.GetSupportedVersions(ctx, req)
}

func (e *endpoint) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	c, err := e.dialIdentity(ctx)
	if err != nil {
		return nil, err
	}
	return c.GetPluginInfo(ctx, req)
}

////////////////////////////////////////////////////////////////////////////////
//                                Node Service                                //
////////////////////////////////////////////////////////////////////////////////

func (e *endpoint) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	c, err := e.dialNode(ctx)
	if err != nil {
		return nil, err
	}
	return c.NodePublishVolume(ctx, req)
}

func (e *endpoint) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	c, err := e.dialNode(ctx)
	if err != nil {
		return nil, err
	}
	return c.NodeUnpublishVolume(ctx, req)
}

func (e *endpoint) GetNodeID(
	ctx context.Context,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {

	c, err := e.dialNode(ctx)
	if err != nil {
		return nil, err
	}
	return c.GetNodeID(ctx, req)
}

func (e *endpoint) ProbeNode(
	ctx context.Context,
	req *csi.ProbeNodeRequest) (
	*csi.ProbeNodeResponse, error) {

	c, err := e.dialNode(ctx)
	if err != nil {
		return nil, err
	}
	return c.ProbeNode(ctx, req)
}

func (e *endpoint) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	c, err := e.dialNode(ctx)
	if err != nil {
		return nil, err
	}
	return c.NodeGetCapabilities(ctx, req)
}

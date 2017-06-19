package main

import (
	"C"

	"fmt"
	"net"
	"os"
	"regexp"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/container-storage-interface/libraries/gocsi/mock/csi"
)
import "log"

////////////////////////////////////////////////////////////////////////////////
//                                 CLI                                        //
////////////////////////////////////////////////////////////////////////////////

// main is ignored when this package is built as a go plug-in
func main() {
	protoAddr := os.Getenv("CSI_ENDPOINT")
	if protoAddr == "" {
		protoAddr = "tcp://127.0.0.1:8080"
	}
	proto, addr, err := parseProtoAddr(protoAddr)
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "error: invalid endpoint: %s\n", protoAddr)
		os.Exit(1)
	}
	l, err := net.Listen(proto, addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to listen: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	e := &endpoint{}
	if err := e.Init(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: endpoint init failed: %v\n", err)
		os.Exit(1)
	}

	if err := e.Serve(ctx, l); err != nil {
		fmt.Fprintf(os.Stderr, "error: grpc failed: %v\n", err)
		os.Exit(1)
	}
}

var addrRX = regexp.MustCompile(
	`(?i)^((?:(?:tcp|udp|ip)[46]?)|(?:unix(?:gram|packet)?))://(.+)$`)

func parseProtoAddr(protoAddr string) (proto string, addr string, err error) {
	m := addrRX.FindStringSubmatch(protoAddr)
	if m == nil {
		return "", "", fmt.Errorf("invalid address: %v", protoAddr)
	}
	return m[1], m[2], nil
}

////////////////////////////////////////////////////////////////////////////////
//                              Go Plug-in                                    //
////////////////////////////////////////////////////////////////////////////////

// Endpoints is an exported symbol that provides a host program
// with a map of the endpoint provider names and constructors.
var Endpoints = map[string]func() interface{}{
	"mock": func() interface{} { return &endpoint{} },
}

type endpoint struct{}

type listVolResult struct{}

func (v *listVolResult) Error() string {
	return ""
}
func (v *listVolResult) Data() []byte {
	r := &csi.ListVolumesResponse{
		Reply: &csi.ListVolumesResponse_Result_{
			Result: &csi.ListVolumesResponse_Result{
				Entries: []*csi.ListVolumesResponse_Result_Entry{
					&csi.ListVolumesResponse_Result_Entry{VolumeInfo: volInfos[0]},
					&csi.ListVolumesResponse_Result_Entry{VolumeInfo: volInfos[1]},
					&csi.ListVolumesResponse_Result_Entry{VolumeInfo: volInfos[2]},
				},
			},
		},
	}
	_ = r
	return nil
}

// Endpoint.Init
func (e *endpoint) Init(ctx context.Context) error {
	log.Println("mock.Init")
	return nil
}

// Endpoint.Serve
func (e *endpoint) Serve(ctx context.Context, li net.Listener) error {
	grpcServer := grpc.NewServer()
	csi.RegisterControllerServer(grpcServer, e)
	csi.RegisterIdentityServer(grpcServer, e)
	csi.RegisterNodeServer(grpcServer, e)
	log.Println("mock.Serve")
	return grpcServer.Serve(li)
}

//  Endpoint.Shutdown
func (e *endpoint) Shutdown(ctx context.Context) error {
	log.Println("mock.Shutdown")
	return nil
}

////////////////////////////////////////////////////////////////////////////////
//                            Controller Service                              //
////////////////////////////////////////////////////////////////////////////////

func (e *endpoint) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	return nil, nil
}

func (e *endpoint) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	return nil, nil
}

func (e *endpoint) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	return nil, nil
}

func (e *endpoint) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	return nil, nil
}

func (e *endpoint) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	return nil, nil
}

func (e *endpoint) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	log.Printf("mock.ListVolumes StartingToken=%v\n", req.GetStartingToken())

	return &csi.ListVolumesResponse{
		Reply: &csi.ListVolumesResponse_Result_{
			Result: &csi.ListVolumesResponse_Result{
				Entries: []*csi.ListVolumesResponse_Result_Entry{
					&csi.ListVolumesResponse_Result_Entry{VolumeInfo: volInfos[0]},
					&csi.ListVolumesResponse_Result_Entry{VolumeInfo: volInfos[1]},
					&csi.ListVolumesResponse_Result_Entry{VolumeInfo: volInfos[2]},
				},
			},
		},
	}, nil
}

func (e *endpoint) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	return nil, nil
}

func (e *endpoint) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
//                             Identity Service                               //
////////////////////////////////////////////////////////////////////////////////

func (e *endpoint) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	return nil, nil
}

func (e *endpoint) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
//                                Node Service                                //
////////////////////////////////////////////////////////////////////////////////

func (e *endpoint) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	return nil, nil
}

func (e *endpoint) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	return nil, nil
}

func (e *endpoint) GetNodeID(
	ctx context.Context,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {

	return nil, nil
}

func (e *endpoint) ProbeNode(
	ctx context.Context,
	req *csi.ProbeNodeRequest) (
	*csi.ProbeNodeResponse, error) {

	return nil, nil
}

func (e *endpoint) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	return nil, nil
}

var volInfos = []*csi.VolumeInfo{
	&csi.VolumeInfo{
		Id: &csi.VolumeID{
			Values: map[string]string{"id": "vol-001"},
		},
	},
	&csi.VolumeInfo{
		Id: &csi.VolumeID{
			Values: map[string]string{"id": "vol-002"},
		},
	},
	&csi.VolumeInfo{
		Id: &csi.VolumeID{
			Values: map[string]string{"id": "vol-003"},
		},
	},
}

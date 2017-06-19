package gocsi_test

import (
	"context"
	"testing"

	"github.com/container-storage-interface/libraries/gocsi"
	"github.com/container-storage-interface/libraries/gocsi/csi"
)

func TestGoCSI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSI")
}

var _ = Describe("CSI", func() {
	var (
		ctx context.Context
		svc gocsi.Service
	)

	BeforeEach(func() {
		var err error
		ctx = context.Background()

		err = gocsi.Init(ctx)
		Ω(err).ShouldNot(HaveOccurred())

		svc, err = gocsi.New(ctx, "mock")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(svc).ShouldNot(BeNil())

		go func() {
			svc.Serve(ctx, nil)
		}()
	})

	AfterEach(func() {
		svc.Shutdown(ctx)
	})

	Context("Controller", func() {
		It("Should list volumes successfully", func() {
			res, err := svc.ListVolumes(
				ctx, &csi.ListVolumesRequest{StartingToken: "1"})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(res).ShouldNot(BeNil())
			Ω(res).Should(Equal(listVolumesResponse))
		})
	})
})

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

var listVolumesResponse = &csi.ListVolumesResponse{
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

package server_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetVersion(t *testing.T) {
	g := NewGomegaWithT(t)
	c, _ := makeGRPCServer(k8sEnv.Rest, t)
	_, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})

	g.Expect(err).NotTo(HaveOccurred())

	resp, err := c.GetVersion(context.Background(), &pb.GetVersionRequest{})

	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(resp.Semver).To(Equal("v0.0.0"))
}

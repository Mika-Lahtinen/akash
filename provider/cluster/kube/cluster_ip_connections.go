package kube

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/ovrclk/akash/manifest"
	akashtypes "github.com/ovrclk/akash/pkg/apis/akash.network/v1"
	"github.com/ovrclk/akash/provider/cluster/kube/builder"
	mtypes "github.com/ovrclk/akash/x/market/types/v1beta2"
	"io"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *client) DeclareIP(ctx context.Context, lID mtypes.LeaseID, serviceName string, externalPort uint32, proto manifest.ServiceProtocol, sharingKey string) error {
	h := sha256.New()
	_, err := io.WriteString(h, lID.String())
	if err != nil {
		return err
	}
	leaseIDHash := h.Sum(nil)
	resourceName := fmt.Sprintf("%X-%s-%d", leaseIDHash, serviceName, externalPort)

	labels := map[string]string{
		builder.AkashManagedLabelName: "true",
	}
	builder.AppendLeaseLabels(lID, labels)
	foundEntry, err := c.ac.AkashV1().ProviderLeasedIPs(c.ns).Get(ctx, resourceName, metav1.GetOptions{})

	var resourceVersion string
	exists := false
	if err != nil {
		if kubeErrors.IsNotFound(err) {
			exists = false
		} else {
			return err
		}
	} else {
		resourceVersion = foundEntry.ObjectMeta.ResourceVersion
	}

	obj := akashtypes.ProviderLeasedIP{
		ObjectMeta: metav1.ObjectMeta{
			Name:            resourceName,
			Labels:          labels,
			ResourceVersion: resourceVersion,
		},
		Spec: akashtypes.ProviderLeasedIPSpec{
			LeaseID:      akashtypes.LeaseIDFromAkash(lID),
			ServiceName:  serviceName,
			ExternalPort: externalPort,
			SharingKey:   sharingKey,
		},
		Status: akashtypes.ProviderLeasedIPStatus{},
	}

	c.log.Info("declaring leased ip", "lease", lID, "service-name", serviceName, "external-port", externalPort, "sharing-key", sharingKey)
	// Create or update the entry
	if exists {
		_, err = c.ac.AkashV1().ProviderLeasedIPs(c.ns).Update(ctx, &obj, metav1.UpdateOptions{})
	} else {
		_, err = c.ac.AkashV1().ProviderLeasedIPs(c.ns).Create(ctx, &obj, metav1.CreateOptions{})
	}
	return err
}
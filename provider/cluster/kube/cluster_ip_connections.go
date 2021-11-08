package kube

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	akashtypes "github.com/ovrclk/akash/pkg/apis/akash.network/v1"
	"github.com/ovrclk/akash/provider/cluster/kube/builder"
	ctypes "github.com/ovrclk/akash/provider/cluster/types"
	mtypes "github.com/ovrclk/akash/x/market/types/v1beta2"
	"io"
	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/pager"
	"strings"
)

func (c *client) DeclareIP(ctx context.Context, lID mtypes.LeaseID,  serviceName string, externalPort uint32, sharingKey string) error {
	h := sha256.New()
	_, err := io.WriteString(h, lID.String())
	if err != nil {
		return err
	}
	leaseIDHash := h.Sum(nil)
	resourceName := fmt.Sprintf("%x-%s-%d", leaseIDHash, serviceName, externalPort)

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
		obj.ResourceVersion = ""
		_, err = c.ac.AkashV1().ProviderLeasedIPs(c.ns).Create(ctx, &obj, metav1.CreateOptions{})
	}

	return err
}

func (c *client) PurgeDeclaredIPs(ctx context.Context, lID mtypes.LeaseID) error {
	labelSelector := &strings.Builder{}
	kubeSelectorForLease(labelSelector, lID)
	result := c.ac.AkashV1().ProviderLeasedIPs(c.ns).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})

	return result
}

func (c *client) ObserveIPState(ctx context.Context) (<-chan ctypes.IPResourceEvent, error) {
	var lastResourceVersion string
	phpager := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
		resources, err := c.ac.AkashV1().ProviderLeasedIPs(c.ns).List(ctx, opts)

		if err == nil && len(resources.GetResourceVersion()) != 0 {
			lastResourceVersion = resources.GetResourceVersion()
		}
		return resources, err
	})

	data := make([]akashtypes.ProviderLeasedIP, 0, 128)
	err := phpager.EachListItem(ctx, metav1.ListOptions{}, func(obj runtime.Object) error {
		plip := obj.(*akashtypes.ProviderLeasedIP)
		data = append(data, *plip)
		return nil
	})

	if err != nil {
		return nil, err
	}

	c.log.Info("starting ip passthrough watch", "resourceVersion", lastResourceVersion)
	watcher, err := c.ac.AkashV1().ProviderLeasedIPs(c.ns).Watch(ctx, metav1.ListOptions{
		TypeMeta:             metav1.TypeMeta{},
		LabelSelector:        "",
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      lastResourceVersion,
		ResourceVersionMatch: "",
		TimeoutSeconds:       nil,
		Limit:                0,
		Continue:             "",
	})
	if err != nil {
		return nil, err
	}

	evData := make([]ipResourceEvent, len(data))
	for i, v := range data {
		ownerAddr, err := sdktypes.AccAddressFromBech32(v.Spec.LeaseID.Owner)
		if err != nil {
			return nil, err
		}
		providerAddr, err := sdktypes.AccAddressFromBech32(v.Spec.LeaseID.Provider)
		if err != nil {
			return nil, err
		}

		leaseID, err := v.Spec.LeaseID.ToAkash()
		if err != nil {
			return nil, err
		}

		ev := ipResourceEvent{
			eventType:    ctypes.ProviderResourceAdd,
			lID: leaseID,
			serviceName:  v.Spec.ServiceName,
			externalPort: v.Spec.ExternalPort,
			ownerAddr: ownerAddr,
			providerAddr: providerAddr,
			sharingKey: v.Spec.SharingKey,
		}
		evData[i] = ev
	}

	data = nil

	output := make(chan ctypes.IPResourceEvent)

	go func() {
		defer close(output)
		for _, v := range evData {
			output <- v
		}
		evData = nil // do not hold the reference

		results := watcher.ResultChan()
		for {
			select {
			case result, ok := <-results:
				if !ok { // Channel closed when an error happens
					return
				}
				plip := result.Object.(*akashtypes.ProviderLeasedIP)
				ownerAddr, err := sdktypes.AccAddressFromBech32(plip.Spec.LeaseID.Owner)
				if err != nil {
					c.log.Error("invalid owner address in provider host", "addr", plip.Spec.LeaseID.Owner, "err", err)
					continue // Ignore event
				}
				providerAddr, err := sdktypes.AccAddressFromBech32(plip.Spec.LeaseID.Provider)
				if err != nil {
					c.log.Error("invalid provider address in provider host", "addr", plip.Spec.LeaseID.Provider, "err", err)
					continue // Ignore event
				}
			    leaseID, err := plip.Spec.LeaseID.ToAkash()
			    if err != nil {
			    	c.log.Error("invalid lease ID", "err", err)
					continue // Ignore event
				}
				ev := ipResourceEvent{
					lID: leaseID,
					ownerAddr:        ownerAddr,
					providerAddr:     providerAddr,
					serviceName:  plip.Spec.ServiceName,
					externalPort: plip.Spec.ExternalPort,
				}
				switch result.Type {

				case watch.Added:
					ev.eventType = ctypes.ProviderResourceAdd
				case watch.Modified:
					ev.eventType = ctypes.ProviderResourceUpdate
				case watch.Deleted:
					ev.eventType = ctypes.ProviderResourceDelete

				case watch.Error:
					// Based on examination of the implementation code, this is basically never called anyways
					c.log.Error("watch error", "err", result.Object)

				default:
					continue
				}

				output <- ev

			case <-ctx.Done():
				return
			}
		}
	}()

	return output, nil
}

type ipResourceEvent struct {
	lID mtypes.LeaseID
	eventType ctypes.ProviderResourceEvent
	serviceName string
	externalPort uint32
	sharingKey string
	providerAddr sdktypes.Address
	ownerAddr sdktypes.Address
}

func (ev ipResourceEvent) GetLeaseID() mtypes.LeaseID {
	return ev.lID
}


func (ev ipResourceEvent) GetEventType() ctypes.ProviderResourceEvent {
	return ev.eventType
}

func (ev ipResourceEvent) GetServiceName() string {
	return ev.serviceName
}

func (ev ipResourceEvent) GetExternalPort() uint32 {
	return ev.externalPort
}

func (ev ipResourceEvent) GetSharingKey() string {
	return ev.sharingKey
}

func (c *client) CreateIPPassthrough(ctx context.Context, lID mtypes.LeaseID, directive ctypes.ClusterIPPassthroughDirective) error {

	ns := builder.LidNS(lID)
	labels := make(map[string]string)
	builder.AppendLeaseLabels(lID, labels)

	selector := map[string]string {
		builder.AkashManagedLabelName: "true",
		builder.AkashManifestServiceLabelName: directive.ServiceName,
	}
	// TODO - specify metallb.universe.tf/address-pool annotation if configured to do so only that pool is used at any time
	annotations := map[string]string {
		"metallb.universe.tf/allow-shared-ip": directive.SharingKey,
	}

	for _, proto := range []corev1.Protocol{
		corev1.ProtocolTCP,
	    corev1.ProtocolUDP, } {

		portName := strings.ToLower(fmt.Sprintf("%s-%d-%v", directive.ServiceName, directive.ServicePort, proto))


		port := corev1.ServicePort{
			Name:        portName,
			Protocol:    proto,
			Port:        int32(directive.ServicePort),
			TargetPort:  intstr.FromInt(int(directive.ServicePort)),
		}

		svc := corev1.Service{

			ObjectMeta: metav1.ObjectMeta{
				Name:                       portName,
				Labels: labels,
				Annotations: annotations,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					port,
				},
				Selector:                      selector,
				Type:                          corev1.ServiceTypeLoadBalancer,
			},
			Status: corev1.ServiceStatus{},
		}

		// TODO - change this to be create or update logic
		_, err := c.kc.CoreV1().Services(ns).Create(ctx, &svc, metav1.CreateOptions{})
		// TODO - handle if some services were already created and this fails
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) PurgeIPPassthrough(ctx context.Context, lID mtypes.LeaseID, serviceName string, externalPort uint32) error {
	return errors.New("nope")
}
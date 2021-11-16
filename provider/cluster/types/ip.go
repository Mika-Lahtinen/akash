package cluster

import (
	"github.com/ovrclk/akash/manifest"
	mtypes "github.com/ovrclk/akash/x/market/types/v1beta2"
)

type IPResourceEvent interface {
	GetLeaseID() mtypes.LeaseID
	GetServiceName() string
	GetExternalPort() uint32
	GetSharingKey() string
	GetProtocol() manifest.ServiceProtocol
	GetEventType() ProviderResourceEvent
}

type IPPassthrough interface {
	GetLeaseID() mtypes.LeaseID
	GetServiceName() string
	GetExternalPort() uint32
	GetSharingKey() string
	GetProtocol() manifest.ServiceProtocol
}

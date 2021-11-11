package cluster

import (
	mtypes "github.com/ovrclk/akash/x/market/types/v1beta2"
)

type IPResourceEvent interface {
	GetLeaseID() mtypes.LeaseID
	GetEventType() ProviderResourceEvent
	GetServiceName() string
	GetExternalPort() uint32
	GetSharingKey() string
}


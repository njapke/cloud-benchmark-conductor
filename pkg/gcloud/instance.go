package gcloud

import "google.golang.org/api/compute/v1"

type Instance struct {
	internalInstance *compute.Instance
}

func (i *Instance) Name() string {
	return i.internalInstance.Name
}

func (i *Instance) ExternalIP() string {
	return i.internalInstance.NetworkInterfaces[0].AccessConfigs[0].NatIP
}

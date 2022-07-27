package gcloud

import (
	"context"
	"net"
	"time"

	"google.golang.org/api/compute/v1"
)

type Instance struct {
	internalInstance *compute.Instance
}

func (i *Instance) Name() string {
	return i.internalInstance.Name
}

func (i *Instance) ExternalIP() string {
	return i.internalInstance.NetworkInterfaces[0].AccessConfigs[0].NatIP
}

func (i *Instance) WaitForSSHPortReady(ctx context.Context) error {
	publicSSHEndpoint := i.ExternalIP() + ":22"
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			conn, err := net.DialTimeout("tcp", publicSSHEndpoint, time.Second)
			if err == nil {
				_ = conn.Close()
				return nil
			}
		}
	}
}

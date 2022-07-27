package gcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

const toolName = "cloud-benchmark-conductor"

func prefixName(n string) string {
	return fmt.Sprintf("cloud-benchmark-conductor-%s", n)
}

type Service struct {
	config         *config.ConductorConfig
	computeService *compute.Service
	crmService     *cloudresourcemanager.Service
	projectId      int64
}

func (s *Service) networkTags() []string {
	return []string{toolName}
}

func (s *Service) labels() map[string]string {
	return map[string]string{
		toolName: "true",
	}
}

func (s *Service) metadata() *compute.Metadata {
	value := fmt.Sprintf("ubuntu:%s ubuntu", s.config.SSHPublicKey)
	return &compute.Metadata{
		Items: []*compute.MetadataItems{
			{Key: "ssh-keys", Value: &value},
		},
	}
}

func (s *Service) getDefaultServiceAccount() string {
	return fmt.Sprintf("%d-compute@developer.gserviceaccount.com", s.projectId)
}

func NewService(conf *config.ConductorConfig) (*Service, error) {
	ctx := context.Background()
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}
	// resolve project id to project number
	crmProject, err := crmService.Projects.Get(conf.Project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", conf.Project, err)
	}

	computeService, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}

	s := &Service{
		config:         conf,
		computeService: computeService,
		crmService:     crmService,
		projectId:      crmProject.ProjectNumber,
	}
	return s, nil
}

func (s *Service) getLatestUbuntuImage(ctx context.Context) (string, error) {
	latestUbuntu, err := s.computeService.Images.
		GetFromFamily("ubuntu-os-cloud", "ubuntu-2204-lts").
		Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return latestUbuntu.SelfLink, nil
}

func (s *Service) waitForOp(ctx context.Context, initialOp *compute.Operation) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			var op *compute.Operation
			var err error
			if initialOp.Zone != "" {
				op, err = s.computeService.ZoneOperations.Get(s.config.Project, s.config.Zone, initialOp.Name).Context(ctx).Do()
			} else if initialOp.Region != "" {
				op, err = s.computeService.RegionOperations.Get(s.config.Project, s.config.Region, initialOp.Name).Context(ctx).Do()
			} else {
				op, err = s.computeService.GlobalOperations.Get(s.config.Project, initialOp.Name).Context(ctx).Do()
			}
			if err != nil {
				return err
			}
			if op.Error != nil && len(op.Error.Errors) > 0 {
				skip := false
				var combinedErrors error
				for _, opErr := range op.Error.Errors {
					if opErr.Code == "RESOURCE_NOT_READY" {
						skip = true
						break
					}
					combinedErrors = multierror.Append(combinedErrors, fmt.Errorf("%s: %s", opErr.Code, opErr.Message))
				}
				if !skip {
					return combinedErrors
				}
			}
			if op.Status == "PENDING" || op.Status == "RUNNING" {
				continue
			}
			if op.Status == "DONE" {
				return nil
			}
			return fmt.Errorf("unknown operation status: %s", op.Status)
		}
	}
}

func (s *Service) GetInstance(ctx context.Context, name string) (*Instance, error) {
	instance, err := s.computeService.Instances.Get(s.config.Project, s.config.Zone, name).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return &Instance{
		internalInstance: instance,
	}, nil
}

func (s *Service) CreateInstance(ctx context.Context, instanceName string) (*Instance, error) {
	latestUbuntu, err := s.getLatestUbuntuImage(ctx)
	if err != nil {
		return nil, err
	}
	machineType := fmt.Sprintf("zones/%s/machineTypes/%s", s.config.Zone, s.config.InstanceType)
	diskType := fmt.Sprintf("zones/%s/diskTypes/pd-balanced", s.config.Zone)
	name := prefixName(instanceName)
	instance := &compute.Instance{
		Name:        name,
		MachineType: machineType,
		Disks: []*compute.AttachedDisk{
			{
				DeviceName: name,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskSizeGb:  10,
					SourceImage: latestUbuntu,
					DiskType:    diskType,
				},
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{
					{
						Name:        "External NAT",
						NetworkTier: "STANDARD",
					},
				},
				Network: "global/networks/default",
			},
		},
		Tags: &compute.Tags{
			Items: s.networkTags(),
		},
		Labels:   s.labels(),
		Metadata: s.metadata(),
		ServiceAccounts: []*compute.ServiceAccount{
			{
				Email:  s.getDefaultServiceAccount(),
				Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"}, // full access
			},
		},
	}
	insertOp, err := s.computeService.Instances.Insert(s.config.Project, s.config.Zone, instance).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	if err := s.waitForOp(ctx, insertOp); err != nil {
		return nil, err
	}
	return s.GetInstance(ctx, instanceName)
}

func (s *Service) DeleteInstances(ctx context.Context) ([]string, error) {
	instancesList, err := s.computeService.Instances.List(s.config.Project, s.config.Zone).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	var mErr error
	deletedInstances := make([]string, 0)
	for _, instance := range instancesList.Items {
		if instance.Labels[toolName] == "true" {
			_, err := s.computeService.Instances.Delete(s.config.Project, s.config.Zone, instance.Name).Context(ctx).Do()
			if err != nil {
				mErr = multierror.Append(mErr, err)
			} else {
				deletedInstances = append(deletedInstances, instance.SelfLink)
			}
		}
	}
	return deletedInstances, mErr
}

func (s *Service) EnsureFirewallRules(ctx context.Context) error {
	_, err := s.computeService.Firewalls.Get(s.config.Project, toolName).Context(ctx).Do()
	if err == nil {
		// firewall already exists, nothing to do
		return nil
	}
	if gErr, ok := err.(*googleapi.Error); !ok || gErr.Code != 404 {
		return err
	}
	firewall := &compute.Firewall{
		Name:         toolName,
		Network:      "global/networks/default",
		Direction:    "INGRESS",
		Priority:     1000,
		TargetTags:   []string{toolName},
		Allowed:      []*compute.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"22"}}},
		SourceRanges: []string{"0.0.0.0/0"},
	}
	insertOp, err := s.computeService.Firewalls.Insert(s.config.Project, firewall).Context(ctx).Do()
	if err != nil {
		return err
	}
	return s.waitForOp(ctx, insertOp)
}

// Cleanup deletes all instances and firewall rules created by this service
func (s *Service) Cleanup(ctx context.Context) ([]string, error) {
	var mErr error
	deletedResources := make([]string, 0)
	if deletedInstances, err := s.DeleteInstances(ctx); err != nil {
		mErr = multierror.Append(mErr, err)
	} else {
		deletedResources = append(deletedResources, deletedInstances...)
	}

	_, err := s.computeService.Firewalls.Delete(s.config.Project, toolName).Context(ctx).Do()
	if err == nil {
		deletedResources = append(deletedResources, fmt.Sprintf("projects/%s/global/firewalls/%s", s.config.Project, toolName))
		return deletedResources, mErr
	}
	// ignore 404 errors, as the firewall may not exist
	if gErr, ok := err.(*googleapi.Error); !ok || gErr.Code != 404 {
		mErr = multierror.Append(mErr, err)
	}
	return deletedResources, mErr
}

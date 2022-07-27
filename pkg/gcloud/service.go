package gcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

var ErrMissingRegion = fmt.Errorf("missing region")
var ErrMissingZone = fmt.Errorf("missing zone")
var ErrMissingProject = fmt.Errorf("missing project")

var allowSSHTag = "allow-ssh"

func prefixName(n string) string {
	return fmt.Sprintf("cbutil-%s", n)
}

type Service struct {
	computeService        *compute.Service
	crmService            *cloudresourcemanager.Service
	project, region, zone string
	SSHPublicKey          string
	projectId             int64
}

func (s *Service) networkTags() []string {
	return []string{"cbutil", allowSSHTag}
}

func (s *Service) labels() map[string]string {
	return map[string]string{
		"cbutil": "true",
	}
}

func (s *Service) metadata() *compute.Metadata {
	value := fmt.Sprintf("ubuntu:%s ubuntu", s.SSHPublicKey)
	return &compute.Metadata{
		Items: []*compute.MetadataItems{
			{Key: "ssh-keys", Value: &value},
		},
	}
}

func (s *Service) getDefaultServiceAccount() string {
	return fmt.Sprintf("%d-compute@developer.gserviceaccount.com", s.projectId)
}

func NewService(project, region, zone string) (*Service, error) {
	var paramsErr error
	if project == "" {
		paramsErr = multierror.Append(paramsErr, ErrMissingProject)
	}
	if region == "" {
		paramsErr = multierror.Append(paramsErr, ErrMissingRegion)
	}
	if zone == "" {
		paramsErr = multierror.Append(paramsErr, ErrMissingZone)
	}
	if paramsErr != nil {
		return nil, paramsErr
	}

	ctx := context.Background()
	computeService, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}

	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}
	crmProject, err := crmService.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", project, err)
	}

	s := &Service{
		computeService: computeService,
		crmService:     crmService,
		project:        project,
		region:         region,
		zone:           zone,
		SSHPublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILjK/2xuvQ0rWgo4FTxUs1reSlvv6+WQC8q3dlNzDCMb",
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
				op, err = s.computeService.ZoneOperations.Get(s.project, s.zone, initialOp.Name).Context(ctx).Do()
			} else if initialOp.Region != "" {
				op, err = s.computeService.RegionOperations.Get(s.project, s.region, initialOp.Name).Context(ctx).Do()
			} else {
				op, err = s.computeService.GlobalOperations.Get(s.project, initialOp.Name).Context(ctx).Do()
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
	instance, err := s.computeService.Instances.Get(s.project, s.zone, name).Context(ctx).Do()
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
	machineType := fmt.Sprintf("zones/%s/machineTypes/f1-micro", s.zone)
	diskType := fmt.Sprintf("zones/%s/diskTypes/pd-balanced", s.zone)
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
	insertOp, err := s.computeService.Instances.Insert(s.project, s.zone, instance).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	if err := s.waitForOp(ctx, insertOp); err != nil {
		return nil, err
	}
	return s.GetInstance(ctx, instanceName)
}

func (s *Service) DeleteInstances(ctx context.Context) ([]string, error) {
	instancesList, err := s.computeService.Instances.List(s.project, s.zone).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	var mErr error
	deletedInstances := make([]string, 0)
	for _, instance := range instancesList.Items {
		if instance.Labels["cbutil"] == "true" {
			_, err := s.computeService.Instances.Delete(s.project, s.zone, instance.Name).Context(ctx).Do()
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
	firewallName := prefixName(allowSSHTag)
	_, err := s.computeService.Firewalls.Get(s.project, firewallName).Context(ctx).Do()
	if err == nil {
		// firewall already exists, nothing to do
		return nil
	}
	if gErr, ok := err.(*googleapi.Error); !ok || gErr.Code != 404 {
		return err
	}
	firewall := &compute.Firewall{
		Name:      firewallName,
		Network:   "global/networks/default",
		Direction: "INGRESS",
		Priority:  1000,
		TargetTags: []string{
			allowSSHTag,
		},
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      []string{"22"},
			},
		},
		SourceRanges: []string{"0.0.0.0/0"},
	}
	inserOp, err := s.computeService.Firewalls.Insert(s.project, firewall).Context(ctx).Do()
	if err != nil {
		return err
	}
	return s.waitForOp(ctx, inserOp)
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

	firewallName := prefixName(allowSSHTag)
	_, err := s.computeService.Firewalls.Delete(s.project, firewallName).Context(ctx).Do()
	if err == nil {
		deletedResources = append(deletedResources, fmt.Sprintf("projects/%s/global/firewalls/%s", s.project, firewallName))
		return deletedResources, mErr
	}
	// ignore 404 errors, as the firewall may not exist
	if gErr, ok := err.(*googleapi.Error); !ok || gErr.Code != 404 {
		mErr = multierror.Append(mErr, err)
	}
	return deletedResources, mErr
}

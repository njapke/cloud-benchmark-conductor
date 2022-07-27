package gcloud

import (
	"context"
	"fmt"

	"google.golang.org/api/compute/v1"
)

var ErrMissingRegion = fmt.Errorf("missing region")
var ErrMissingZone = fmt.Errorf("missing zone")
var ErrMissingProject = fmt.Errorf("missing project")

type Service struct {
	computeService *compute.Service
	Region         string
	Zone           string
	Project        string
	SSHPublicKey   string
}

func (s *Service) NetworkTags() []string {
	return []string{"cbutil", "allow-ssh"}
}

func (s *Service) Labels() map[string]string {
	return map[string]string{
		"cbutil": "true",
	}
}

func (s *Service) Metadata() *compute.Metadata {
	value := fmt.Sprintf("ubuntu:%s ubuntu", s.SSHPublicKey)
	return &compute.Metadata{
		Items: []*compute.MetadataItems{
			{Key: "ssh-keys", Value: &value},
		},
	}
}

func NewService(project, region, zone string) (*Service, error) {
	if project == "" {
		return nil, ErrMissingRegion
	}
	if region == "" {
		return nil, ErrMissingProject
	}
	if zone == "" {
		return nil, ErrMissingZone
	}
	computeService, err := compute.NewService(context.Background())
	if err != nil {
		return nil, err
	}
	s := &Service{
		computeService: computeService,
		Project:        project,
		Region:         region,
		Zone:           zone,
		SSHPublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILjK/2xuvQ0rWgo4FTxUs1reSlvv6+WQC8q3dlNzDCMb",
	}

	return s, nil
}

func (s *Service) GetLatestUbuntuImage(ctx context.Context) (string, error) {
	latestUbuntu, err := s.computeService.Images.
		GetFromFamily("ubuntu-os-cloud", "ubuntu-2204-lts").
		Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return latestUbuntu.SelfLink, nil
}

// https://cloud.google.com/compute/docs/instances/create-start-instance#create_a_vm_instance_in_a_specific_subnet
func (s *Service) CreateInstance(ctx context.Context, instanceName string) error {
	latestUbuntu, err := s.GetLatestUbuntuImage(ctx)
	if err != nil {
		return err
	}
	machineType := fmt.Sprintf("zones/%s/machineTypes/f1-micro", s.Zone)
	diskType := fmt.Sprintf("zones/%s/diskTypes/pd-balanced", s.Zone)
	instance := &compute.Instance{
		Name:        instanceName,
		MachineType: machineType,
		Disks: []*compute.AttachedDisk{
			{
				DeviceName: instanceName,
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
			Items: s.NetworkTags(),
		},
		Labels:   s.Labels(),
		Metadata: s.Metadata(),
	}
	insertOp, err := s.computeService.Instances.Insert(s.Project, s.Zone, instance).Context(ctx).Do()
	if err != nil {
		return err
	}
	// https://github.com/hashicorp/terraform-provider-google/blob/09ffd42bd78b8d26f78f511336351eddbb7d1cee/google/compute_operation.go#L56
	//s.computeService.ZoneOperations.Get(s.Project, s.Zone, insertOp.Name).Do()
	fmt.Printf("%#v", insertOp)
	return nil
}

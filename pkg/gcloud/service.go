package gcloud

import (
	"context"
	"errors"
	"fmt"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	resourcemanagerpb "google.golang.org/genproto/googleapis/cloud/resourcemanager/v3"

	"github.com/christophwitzko/masters-thesis/pkg/config"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/crypto/ssh"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

const (
	toolName       = "cloud-benchmark-conductor"
	instancePrefix = "cbc-"
)

func prefixName(n string) string {
	return instancePrefix + trimPrefixName(n)
}

func trimPrefixName(n string) string {
	return strings.TrimPrefix(n, instancePrefix)
}

type Service interface {
	Config() *config.ConductorConfig
	GetInstance(ctx context.Context, name string) (Instance, error)
	GetOrCreateInstance(ctx context.Context, name, instanceType string) (Instance, error)
	EnsureFirewallRules(ctx context.Context) error
	ListInstances(ctx context.Context) ([]string, error)
	DeleteInstance(ctx context.Context, instanceName string) error
	DeleteFirewallRules(ctx context.Context) (bool, error)
	Close() error
}

var _ Service = (*service)(nil)

type service struct {
	config          *config.ConductorConfig
	imagesClient    *compute.ImagesClient
	instancesClient *compute.InstancesClient
	firewallClient  *compute.FirewallsClient
	projectNumber   string
}

func NewService(conf *config.ConductorConfig) (Service, error) {
	ctx := context.Background()
	projectsClient, err := resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		return nil, err
	}
	defer projectsClient.Close()
	// resolve project id to project number
	projectNumber, err := resolveProjectNumber(projectsClient.SearchProjects(ctx,
		&resourcemanagerpb.SearchProjectsRequest{Query: "projectId:" + conf.Project},
	))
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", conf.Project, err)
	}

	imagesClient, err := compute.NewImagesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	firewallClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	s := &service{
		config:          conf,
		imagesClient:    imagesClient,
		instancesClient: instancesClient,
		firewallClient:  firewallClient,
		projectNumber:   projectNumber,
	}
	return s, nil
}

func (s *service) networkTags() []string {
	return []string{toolName}
}

func (s *service) labels() map[string]string {
	return map[string]string{
		toolName: "true",
	}
}

func (s *service) metadata() *computepb.Metadata {
	value := fmt.Sprintf("ubuntu:%s", ssh.MarshalAuthorizedKey(s.config.SSHSigner.PublicKey()))
	return &computepb.Metadata{
		Items: []*computepb.Items{
			{Key: proto.String("ssh-keys"), Value: &value},
		},
	}
}

func (s *service) getDefaultServiceAccount() string {
	return fmt.Sprintf("%s-compute@developer.gserviceaccount.com", s.projectNumber)
}

func resolveProjectNumber(it *resourcemanager.ProjectIterator) (string, error) {
	p, err := it.Next()
	if errors.Is(err, iterator.Done) {
		return "", fmt.Errorf("no project found")
	}
	if err != nil {
		return "", err
	}
	_, projectNumber, _ := strings.Cut(p.GetName(), "/")
	return projectNumber, nil
}

func (s *service) getLatestUbuntuImage(ctx context.Context) (*string, error) {
	latestUbuntu, err := s.imagesClient.GetFromFamily(ctx, &computepb.GetFromFamilyImageRequest{
		Project: "ubuntu-os-cloud",
		Family:  "ubuntu-2204-lts",
	})
	if err != nil {
		return nil, err
	}
	return latestUbuntu.SelfLink, nil
}

// Config returns the global config
func (s *service) Config() *config.ConductorConfig {
	return s.config
}

// GetInstance returns the instance with the given name
func (s *service) GetInstance(ctx context.Context, name string) (Instance, error) {
	internalInstance, err := s.instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Project:  s.config.Project,
		Zone:     s.config.Zone,
		Instance: prefixName(name),
	})
	if err != nil {
		return nil, err
	}
	return &instance{
		config:           s.config,
		internalInstance: internalInstance,
	}, nil
}

// GetOrCreateInstance tries to get an instance with the given name. If it does not exist, it will be created.
func (s *service) GetOrCreateInstance(ctx context.Context, name, instanceType string) (Instance, error) {
	latestUbuntu, err := s.getLatestUbuntuImage(ctx)
	if err != nil {
		return nil, err
	}

	// if instance already exists, return it
	existingInstance, err := s.GetInstance(ctx, name)
	if err == nil {
		return existingInstance, nil
	}
	var gErr *googleapi.Error
	if !errors.As(err, &gErr) || gErr.Code != 404 {
		return nil, err
	}

	prefixedInstanceName := prefixName(name)
	if instanceType == "" {
		instanceType = s.config.DefaultInstanceType
	}
	machineType := fmt.Sprintf("zones/%s/machineTypes/%s", s.config.Zone, instanceType)
	diskType := fmt.Sprintf("zones/%s/diskTypes/pd-balanced", s.config.Zone)
	insertInstance := &computepb.InsertInstanceRequest{
		Project: s.config.Project,
		Zone:    s.config.Zone,
		InstanceResource: &computepb.Instance{
			Name:        &prefixedInstanceName,
			MachineType: &machineType,
			Disks: []*computepb.AttachedDisk{
				{
					DeviceName: &prefixedInstanceName,
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						DiskSizeGb:  proto.Int64(20),
						SourceImage: latestUbuntu,
						DiskType:    &diskType,
					},
					AutoDelete: proto.Bool(true),
					Boot:       proto.Bool(true),
					Type:       proto.String(computepb.AttachedDisk_PERSISTENT.String()),
				},
			},
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					AccessConfigs: []*computepb.AccessConfig{
						{
							Name:        proto.String("External NAT"),
							NetworkTier: proto.String(computepb.AccessConfig_STANDARD.String()),
						},
					},
					Network: proto.String("global/networks/default"),
				},
			},
			Tags: &computepb.Tags{
				Items: s.networkTags(),
			},
			Labels:   s.labels(),
			Metadata: s.metadata(),
			ServiceAccounts: []*computepb.ServiceAccount{
				{
					Email:  proto.String(s.getDefaultServiceAccount()),
					Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"}, // full access
				},
			},
		},
	}
	insertOp, err := s.instancesClient.Insert(ctx, insertInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	if err := insertOp.Wait(ctx); err != nil {
		return nil, fmt.Errorf("failed to wait for operation: %w", err)
	}
	return s.GetInstance(ctx, name)
}

func (s *service) EnsureFirewallRules(ctx context.Context) error {
	_, err := s.firewallClient.Get(ctx, &computepb.GetFirewallRequest{
		Project:  s.config.Project,
		Firewall: toolName,
	})
	if err == nil {
		// firewall already exists, nothing to do
		return nil
	}
	var gErr *googleapi.Error
	if !errors.As(err, &gErr) || gErr.Code != 404 {
		return err
	}

	insertOp, err := s.firewallClient.Insert(ctx, &computepb.InsertFirewallRequest{
		Project: s.config.Project,
		FirewallResource: &computepb.Firewall{
			Name:       proto.String(toolName),
			Network:    proto.String("global/networks/default"),
			Direction:  proto.String(computepb.Firewall_INGRESS.String()),
			Priority:   proto.Int32(1000),
			TargetTags: []string{toolName},
			Allowed: []*computepb.Allowed{{
				IPProtocol: proto.String("tcp"),
				// TODO: remove application ports after testing and use internal networking
				Ports: []string{"22", "3000-3010"},
			}},
			SourceRanges: []string{"0.0.0.0/0"},
		},
	})
	if err != nil {
		return err
	}
	return insertOp.Wait(ctx)
}

func (s *service) DeleteInstance(ctx context.Context, instanceName string) error {
	delOp, err := s.instancesClient.Delete(ctx, &computepb.DeleteInstanceRequest{
		Project:  s.config.Project,
		Zone:     s.config.Zone,
		Instance: prefixName(instanceName),
	})
	if err != nil {
		return err
	}

	err = delOp.Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) ListInstances(ctx context.Context) ([]string, error) {
	instances := s.instancesClient.List(ctx, &computepb.ListInstancesRequest{
		Project: s.config.Project,
		Zone:    s.config.Zone,
		Filter:  proto.String(fmt.Sprintf("labels.%s=true", toolName)),
	})
	instanceNames := make([]string, 0)
	for {
		instance, err := instances.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		instanceNames = append(instanceNames, *instance.Name)
	}
	return instanceNames, nil
}

func (s *service) DeleteFirewallRules(ctx context.Context) (bool, error) {
	_, err := s.firewallClient.Delete(ctx, &computepb.DeleteFirewallRequest{
		Project:  s.config.Project,
		Firewall: toolName,
	})
	if err == nil {
		return true, nil
	}

	// ignore 404 errors, as the firewall may not exist
	var gErr *googleapi.Error
	if !errors.As(err, &gErr) || gErr.Code != 404 {
		return false, err
	}
	return false, nil
}

// Close all api clients
func (s *service) Close() error {
	var mErr error
	if err := s.imagesClient.Close(); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	if err := s.instancesClient.Close(); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	if err := s.firewallClient.Close(); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	return mErr
}

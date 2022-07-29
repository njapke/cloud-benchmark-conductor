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

	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/crypto/ssh"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

const toolName = "cloud-benchmark-conductor"

func prefixName(n string) string {
	return fmt.Sprintf("cloud-benchmark-conductor-%s", n)
}

type Service struct {
	config          *config.ConductorConfig
	imagesClient    *compute.ImagesClient
	instancesClient *compute.InstancesClient
	firewallClient  *compute.FirewallsClient
	projectNumber   string
}

func (s *Service) networkTags() []string {
	return []string{toolName}
}

func (s *Service) labels() map[string]string {
	return map[string]string{
		toolName: "true",
	}
}

func (s *Service) metadata() *computepb.Metadata {
	value := fmt.Sprintf("ubuntu:%s ubuntu", ssh.MarshalAuthorizedKey(s.config.SSHSigner.PublicKey()))
	return &computepb.Metadata{
		Items: []*computepb.Items{
			{Key: proto.String("ssh-keys"), Value: &value},
		},
	}
}

func (s *Service) getDefaultServiceAccount() string {
	return fmt.Sprintf("%s-compute@developer.gserviceaccount.com", s.projectNumber)
}

func resolveProjectNumber(it *resourcemanager.ProjectIterator) (string, error) {
	p, err := it.Next()
	if err == iterator.Done {
		return "", fmt.Errorf("no project found")
	}
	if err != nil {
		return "", err
	}
	_, projectNumber, _ := strings.Cut(p.GetName(), "/")
	return projectNumber, nil
}

func NewService(conf *config.ConductorConfig) (*Service, error) {
	ctx := context.Background()
	projectsClient, err := resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		return nil, err
	}
	defer projectsClient.Close()
	// resolve project id to project number
	projectNumber, err := resolveProjectNumber(projectsClient.SearchProjects(ctx,
		&resourcemanagerpb.SearchProjectsRequest{Query: "name:" + conf.Project},
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
	s := &Service{
		config:          conf,
		imagesClient:    imagesClient,
		instancesClient: instancesClient,
		firewallClient:  firewallClient,
		projectNumber:   projectNumber,
	}
	return s, nil
}

func (s *Service) getLatestUbuntuImage(ctx context.Context) (*string, error) {
	latestUbuntu, err := s.imagesClient.GetFromFamily(ctx, &computepb.GetFromFamilyImageRequest{
		Project: "ubuntu-os-cloud",
		Family:  "ubuntu-2204-lts",
	})
	if err != nil {
		return nil, err
	}
	return latestUbuntu.SelfLink, nil
}

func (s *Service) GetInstance(ctx context.Context, name string) (*Instance, error) {
	instance, err := s.instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Project:  s.config.Project,
		Zone:     s.config.Zone,
		Instance: prefixName(name),
	})
	if err != nil {
		return nil, err
	}
	return &Instance{
		config:           s.config,
		internalInstance: instance,
	}, nil
}

func (s *Service) CreateInstance(ctx context.Context, name string) (*Instance, error) {
	latestUbuntu, err := s.getLatestUbuntuImage(ctx)
	if err != nil {
		return nil, err
	}

	// if instance already exists, return it
	instance, err := s.GetInstance(ctx, name)
	if err == nil {
		return instance, nil
	}
	if gErr, ok := errors.Unwrap(err).(*googleapi.Error); !ok || gErr.Code != 404 {
		return nil, err
	}

	prefixedInstanceName := prefixName(name)
	machineType := fmt.Sprintf("zones/%s/machineTypes/%s", s.config.Zone, s.config.InstanceType)
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

func (s *Service) CleanupInstances(ctx context.Context) ([]string, error) {
	instances := s.instancesClient.List(ctx, &computepb.ListInstancesRequest{
		Project: s.config.Project,
		Zone:    s.config.Zone,
		Filter:  proto.String(fmt.Sprintf("labels.%s=true", toolName)),
	})

	var mErr error
	deletedInstances := make([]string, 0)
	for {
		instance, err := instances.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		instanceName := *instance.Name
		_, err = s.instancesClient.Delete(ctx, &computepb.DeleteInstanceRequest{
			Project:  s.config.Project,
			Zone:     s.config.Zone,
			Instance: instanceName,
		})
		if err != nil {
			mErr = multierror.Append(mErr, err)
		} else {
			deletedInstances = append(deletedInstances, "instances/"+instanceName)
		}
	}
	return deletedInstances, mErr
}

func (s *Service) EnsureFirewallRules(ctx context.Context) error {
	_, err := s.firewallClient.Get(ctx, &computepb.GetFirewallRequest{
		Project:  s.config.Project,
		Firewall: toolName,
	})
	if err == nil {
		// firewall already exists, nothing to do
		return nil
	}
	if gErr, ok := errors.Unwrap(err).(*googleapi.Error); !ok || gErr.Code != 404 {
		return err
	}

	insertOp, err := s.firewallClient.Insert(ctx, &computepb.InsertFirewallRequest{
		Project: s.config.Project,
		FirewallResource: &computepb.Firewall{
			Name:         proto.String(toolName),
			Network:      proto.String("global/networks/default"),
			Direction:    proto.String(computepb.Firewall_INGRESS.String()),
			Priority:     proto.Int32(1000),
			TargetTags:   []string{toolName},
			Allowed:      []*computepb.Allowed{{IPProtocol: proto.String("tcp"), Ports: []string{"22"}}},
			SourceRanges: []string{"0.0.0.0/0"},
		},
	})
	if err != nil {
		return err
	}
	return insertOp.Wait(ctx)
}

// Cleanup deletes all instances and firewall rules created by this service
func (s *Service) Cleanup(ctx context.Context) ([]string, error) {
	var mErr error
	deletedResources := make([]string, 0)
	deletedInstances, err := s.CleanupInstances(ctx)
	if err != nil {
		mErr = multierror.Append(mErr, err)
	}
	if deletedInstances != nil {
		deletedResources = append(deletedResources, deletedInstances...)
	}

	_, err = s.firewallClient.Delete(ctx, &computepb.DeleteFirewallRequest{
		Project:  s.config.Project,
		Firewall: toolName,
	})
	if err == nil {
		deletedResources = append(deletedResources, "firewalls/"+toolName)
		return deletedResources, mErr
	}

	// ignore 404 errors, as the firewall may not exist
	if gErr, ok := errors.Unwrap(err).(*googleapi.Error); !ok || gErr.Code != 404 {
		mErr = multierror.Append(mErr, err)
	}
	return deletedResources, mErr
}

func (s *Service) Close() error {
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

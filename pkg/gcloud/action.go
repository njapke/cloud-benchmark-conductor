package gcloud

import (
	"context"
)

type Action interface {
	Run(ctx context.Context, instance Instance) error
	Name() string
}

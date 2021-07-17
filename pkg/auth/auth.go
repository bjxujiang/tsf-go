package auth

import (
	"context"

	"github.com/bjxujiang/tsf-go/pkg/config"
	"github.com/bjxujiang/tsf-go/pkg/naming"
)

type Builder interface {
	Build(cfg config.Source, svc naming.Service) Auth
}

type Auth interface {
	// api为被访问的接口名
	Verify(ctx context.Context, api string) error
}

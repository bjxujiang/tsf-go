package route

import (
	"context"

	"github.com/bjxujiang/tsf-go/pkg/naming"
)

type Router interface {
	Select(ctx context.Context, svc naming.Service, nodes []naming.Instance) (selects []naming.Instance)
}

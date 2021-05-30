package tsf

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/api/metadata"
	tgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/swagger-api/openapiv2"
	"github.com/tencentyun/tsf-go/balancer/random"
	"github.com/tencentyun/tsf-go/grpc/balancer/multi"
	httpMulti "github.com/tencentyun/tsf-go/http/balancer/multi"
	"github.com/tencentyun/tsf-go/naming/consul"
	"github.com/tencentyun/tsf-go/pkg/sys/env"
	"github.com/tencentyun/tsf-go/pkg/version"
	"github.com/tencentyun/tsf-go/route/composite"
	"google.golang.org/grpc"
)

// ServerOption is HTTP server option.
type ServerOption func(*serverOptions)

func ProtoServiceName(fullname string) ServerOption {
	return func(s *serverOptions) {
		s.protoService = fullname
	}
}

func GRPCServer(srv *grpc.Server) ServerOption {
	return func(s *serverOptions) {
		s.srv = srv
	}
}

func APIMeta(enable bool) ServerOption {
	return func(s *serverOptions) {
		s.apiMeta = enable
	}
}

type serverOptions struct {
	protoService string
	srv          *grpc.Server
	apiMeta      bool
}

func Metadata(optFuncs ...ServerOption) (opt kratos.Option) {
	var opts serverOptions = serverOptions{
		apiMeta: true,
	}
	for _, o := range optFuncs {
		o(&opts)
	}

	md := map[string]string{
		"TSF_APPLICATION_ID": env.ApplicationID(),
		"TSF_GROUP_ID":       env.GroupID(),
		"TSF_INSTNACE_ID":    env.InstanceId(),
		"TSF_PROG_VERSION":   env.ProgVersion(),
		"TSF_ZONE":           env.Zone(),
		"TSF_REGION":         env.Region(),
		"TSF_NAMESPACE_ID":   env.NamespaceID(),
		"TSF_SDK_VERSION":    version.GetHumanVersion(),
	}
	if opts.apiMeta {

		var apiSrv *openapiv2.Service
		if opts.srv != nil {
			apiSrv = openapiv2.New(opts.srv)
		} else {
			apiSrv = openapiv2.New(opts.srv)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		var apiMeta string
		if opts.protoService != "" {
			apiMeta, _ = apiSrv.GetServiceOpenAPI(ctx, &metadata.GetServiceDescRequest{Name: opts.protoService})

		} else {
			reply, err := apiSrv.ListServices(ctx, &metadata.ListServicesRequest{})
			if err == nil {
				for _, service := range reply.Services {
					if service != "grpc.health.v1.Health" && service != "grpc.reflection.v1alpha.ServerReflection" && service != "kratos.api.Metadata" {
						apiMeta, _ = apiSrv.GetServiceOpenAPI(ctx, &metadata.GetServiceDescRequest{Name: service})
						break
					}
				}
			}
		}
		if apiMeta != "" {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
			_, err := zw.Write([]byte(apiMeta))
			if err == nil {
				err = zw.Close()
				if err == nil {
					res := base64.StdEncoding.EncodeToString(buf.Bytes())
					md["TSF_API_METAS"] = res
				}
			}
		}
	}
	opt = kratos.Metadata(md)
	return
}

func ID() kratos.Option {
	return kratos.ID(env.InstanceId())
}
func Registrar() kratos.Option {
	return kratos.Registrar(consul.DefaultConsul())
}

func ClientGrpcOptions() tgrpc.ClientOption {
	// 将wrr负载均衡模块注入至grpc
	router := composite.DefaultComposite()
	multi.Register(router)
	return tgrpc.WithOptions(grpc.WithBalancerName("tsf-random"))
}

func ClientHTTPOptions() http.ClientOption {
	router := composite.DefaultComposite()
	b := &random.Picker{}
	return http.WithBalancer(httpMulti.New(router, b))
}
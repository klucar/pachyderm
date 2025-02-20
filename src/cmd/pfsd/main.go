package main

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/net/context"

	"go.pedge.io/env"
	"go.pedge.io/proto/server"
	"go.pedge.io/protolog"

	"github.com/gengo/grpc-gateway/runtime"
	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/drive/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/netutil"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
	"google.golang.org/grpc"
)

type appEnv struct {
	DriverRoot  string `env:"PFS_DRIVER_ROOT,required"`
	DriverType  string `env:"PFS_DRIVER_TYPE,default=btrfs"`
	NumShards   uint64 `env:"PFS_NUM_SHARDS,default=16"`
	NumReplicas uint64 `env:"PFS_NUM_REPLICAS"`
	Address     string `env:"PFS_ADDRESS"`
	Port        int    `env:"PFS_PORT,default=650"`
	HTTPPort    int    `env:"PFS_HTTP_PORT,default=750"`
	DebugPort   int    `env:"PFS_TRACE_PORT,default=1050"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	discoveryClient, err := getEtcdClient()
	if err != nil {
		return err
	}
	address := appEnv.Address
	if address == "" {
		address, err = netutil.ExternalIP()
		if err != nil {
			return err
		}
	}
	address = fmt.Sprintf("%s:%d", address, appEnv.Port)
	sharder := shard.NewSharder(
		discoveryClient,
		appEnv.NumShards,
		appEnv.NumReplicas,
		"namespace",
	)
	var driver drive.Driver
	switch appEnv.DriverType {
	case "btrfs":
		driver, err = btrfs.NewDriver(appEnv.DriverRoot, "")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown value for PFS_DRIVER_TYPE: %s", appEnv.DriverType)
	}
	apiServer := server.NewAPIServer(
		route.NewSharder(
			appEnv.NumShards,
			0,
		),
		route.NewRouter(
			sharder,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		),
	)
	go func() {
		if err := sharder.RegisterFrontend(nil, address, apiServer); err != nil {
			protolog.Printf("Error from sharder.RegisterFrontend %s", err.Error())
		}
	}()
	internalAPIServer := server.NewInternalAPIServer(
		route.NewSharder(
			appEnv.NumShards,
			0,
		),
		route.NewRouter(
			sharder,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		),
		driver,
	)
	go func() {
		if err := sharder.Register(nil, address, internalAPIServer); err != nil {
			protolog.Printf("Error from sharder.Register %s", err.Error())
		}
	}()
	return protoserver.Serve(
		uint16(appEnv.Port),
		func(s *grpc.Server) {
			pfs.RegisterAPIServer(s, apiServer)
			pfs.RegisterInternalAPIServer(s, internalAPIServer)
		},
		protoserver.ServeOptions{
			HTTPPort:  uint16(appEnv.HTTPPort),
			DebugPort: uint16(appEnv.DebugPort),
			Version:   pachyderm.Version,
			HTTPRegisterFunc: func(ctx context.Context, mux *runtime.ServeMux, clientConn *grpc.ClientConn) error {
				return pfs.RegisterAPIHandler(ctx, mux, clientConn)
			},
		},
	)
}

func getEtcdClient() (discovery.Client, error) {
	etcdAddress, err := getEtcdAddress()
	if err != nil {
		return nil, err
	}
	return discovery.NewEtcdClient(etcdAddress), nil
}

func getEtcdAddress() (string, error) {
	etcdAddr := os.Getenv("ETCD_PORT_2379_TCP_ADDR")
	if etcdAddr == "" {
		return "", errors.New("ETCD_PORT_2379_TCP_ADDR not set")
	}
	return fmt.Sprintf("http://%s:2379", etcdAddr), nil
}

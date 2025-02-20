package main

import (
	"errors"
	"fmt"
	"os"

	"go.pedge.io/env"
	"go.pedge.io/proto/server"
	"go.pedge.io/protolog"
	"google.golang.org/grpc"
	kube "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/jobserver"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	persistserver "github.com/pachyderm/pachyderm/src/pps/persist/server"
	"github.com/pachyderm/pachyderm/src/pps/pipelineserver"
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	PfsAddress         string `env:"PFS_ADDRESS"`
	PfsMountDir        string `env:"PFS_MOUNT_DIR"`
	Address            string `env:"PPS_ADDRESS,default=0.0.0.0"`
	Port               int    `env:"PPS_PORT,default=651"`
	DatabaseAddress    string `env:"PPS_DATABASE_ADDRESS"`
	DatabaseName       string `env:"PPS_DATABASE_NAME,default=pachyderm"`
	DebugPort          int    `env:"PPS_TRACE_PORT,default=1051"`
	RemoveContainers   bool   `env:"PPS_REMOVE_CONTAINERS"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	rethinkAPIClient, err := getRethinkAPIClient(appEnv.DatabaseAddress, appEnv.DatabaseName)
	if err != nil {
		return err
	}
	pfsdAddress, err := getPfsdAddress()
	if err != nil {
		return err
	}
	clientConn, err := grpc.Dial(pfsdAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	pfsAPIClient := pfs.NewAPIClient(clientConn)
	kubeAddr, err := getKubeAddress()
	if err != nil {
		return err
	}
	config := &kube.Config{
		Host:     kubeAddr,
		Insecure: true,
	}
	kubeClient, err := kube.New(config)
	if err != nil {
		protolog.Printf("Error creating kubernetes client: %s", err.Error())
	}
	jobAPIServer := jobserver.NewAPIServer(
		pfsAPIClient,
		rethinkAPIClient,
		kubeClient,
	)
	jobAPIClient := pps.NewLocalJobAPIClient(jobAPIServer)
	pipelineAPIServer := pipelineserver.NewAPIServer(pfsAPIClient, jobAPIClient, rethinkAPIClient)
	if err := pipelineAPIServer.Start(); err != nil {
		return err
	}
	return protoserver.Serve(
		uint16(appEnv.Port),
		func(s *grpc.Server) {
			pps.RegisterJobAPIServer(s, jobAPIServer)
			pps.RegisterInternalJobAPIServer(s, jobAPIServer)
			pps.RegisterPipelineAPIServer(s, pipelineAPIServer)
		},
		protoserver.ServeOptions{
			DebugPort: uint16(appEnv.DebugPort),
			Version:   pachyderm.Version,
		},
	)
}

func getRethinkAPIClient(address string, databaseName string) (persist.APIClient, error) {
	var err error
	if address == "" {
		address, err = getRethinkAddress()
		if err != nil {
			return nil, err
		}
	}
	if err := persistserver.InitDBs(address, databaseName); err != nil {
		return nil, err
	}
	rethinkAPIServer, err := persistserver.NewRethinkAPIServer(address, databaseName)
	if err != nil {
		return nil, err
	}
	return persist.NewLocalAPIClient(rethinkAPIServer), nil
}

func getRethinkAddress() (string, error) {
	rethinkAddr := os.Getenv("RETHINK_PORT_28015_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("RETHINK_PORT_28015_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:28015", rethinkAddr), nil
}

func getPfsdAddress() (string, error) {
	pfsdAddr := os.Getenv("PFSD_PORT_650_TCP_ADDR")
	if pfsdAddr == "" {
		return "", errors.New("PFSD_PORT_650_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:650", pfsdAddr), nil
}

func getKubeAddress() (string, error) {
	kubedAddr := os.Getenv("KUBERNETES_PORT_443_TCP_ADDR")
	if kubedAddr == "" {
		return "", errors.New("KUBERNETES_PORT_443_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:443", kubedAddr), nil
}

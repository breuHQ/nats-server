package shared

// import (
// 	"context"
// 	"path"
// 	"runtime"

// 	"github.com/docker/docker/api/types/container"
// 	"github.com/docker/go-connections/nat"
// 	"github.com/testcontainers/testcontainers-go"
// 	"github.com/testcontainers/testcontainers-go/wait"
// )

// type (
// 	Stack struct {
// 		nats    testcontainers.Container
// 		nozl    testcontainers.Container
// 		network testcontainers.Network
// 		ctx     context.Context
// 	}
// )

// func (s *Stack) SetupStack() {
// 	s.ctx = context.Background()
// 	s.SetupNetwork()
// 	s.SetupContainers()
// }

// func (s *Stack) SetupContainers() {
// 	req := testcontainers.ContainerRequest{
// 		Image:        "nats:2.8.4",
// 		ExposedPorts: []string{"4222/tcp"},
// 		WaitingFor:   wait.ForLog("Server is ready"),
// 		Networks:     []string{"nozl-network"},
// 		Cmd:          []string{"--cluster_name", "nozl", "-js"},
// 		Name:         "nats-io",
// 	}

// 	natsC, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
// 		ContainerRequest: req,
// 		Started:          true,
// 	})

// 	if err != nil {
// 		Logger.Error(err.Error())
// 		return
// 	}

// 	_, caller, _, _ := runtime.Caller(0)
// 	pkgroot := path.Join(path.Dir(caller), "..", "..")

// 	mounts := testcontainers.ContainerMounts{
// 		testcontainers.ContainerMount{
// 			Source: testcontainers.GenericVolumeMountSource{Name: "est--go-pkg-mod"},
// 			Target: "/go/pkg/mod",
// 		},
// 		testcontainers.ContainerMount{
// 			Source: testcontainers.GenericVolumeMountSource{Name: "test--go-build-cache"},
// 			Target: "/root/.cache/go-build",
// 		},
// 		testcontainers.ContainerMount{
// 			Source: testcontainers.GenericBindMountSource{HostPath: pkgroot},
// 			Target: testcontainers.ContainerMountTarget("/test"),
// 		},
// 		testcontainers.ContainerMount{
// 			Source: testcontainers.GenericBindMountSource{HostPath: path.Join(pkgroot, "deploy", "air", "nozl.toml")},
// 			Target: testcontainers.ContainerMountTarget("/test/.air.toml"),
// 		},
// 	}

// 	httpPort, _ := nat.NewPort("tcp", "1323")

// 	req = testcontainers.ContainerRequest{
// 		Image:        "cosmtrek/air:v1.42.0",
// 		ExposedPorts: []string{"1323/tcp"},
// 		Env: map[string]string{
// 			"JWT_SECRET":      "JaNdRgUkXp2s5v8y/B?E(G+KbPeShVmY",
// 			"NATS_SERVER_URL": "nats://nats-io:4222",
// 		},
// 		Networks: []string{"nozl-network"},
// 		Mounts:   mounts,
// 		ConfigModifier: func(config *container.Config) {
// 			config.WorkingDir = "/test"
// 		},
// 		Name: "nozl-api",
// 		WaitingFor: wait.ForListeningPort(httpPort),
// 	}

// 	nozlC, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
// 		ContainerRequest: req,
// 		Started:          true,
// 	})

// 	if err != nil {
// 		Logger.Error(err.Error())
// 		return
// 	}

// 	s.nats = natsC
// 	s.nozl = nozlC
// }

// func (s *Stack) SetupNetwork() {
// 	req := testcontainers.GenericNetworkRequest{
// 		NetworkRequest: testcontainers.NetworkRequest{Name: "nozl-network", CheckDuplicate: true},
// 	}

// 	network, err := testcontainers.GenericNetwork(s.ctx, req)
// 	if err != nil {
// 		Logger.Error(err.Error())
// 		return
// 	}

// 	s.network = network
// }

// func (s *Stack) TeardownStack() {
// 	defer func() {
// 		if err := s.nozl.Terminate(s.ctx); err != nil {
// 			Logger.Error(err.Error())
// 		}

// 		if err := s.nats.Terminate(s.ctx); err != nil {
// 			Logger.Error(err.Error())
// 		}

// 		if err := s.network.Remove(s.ctx); err != nil {
// 			Logger.Error(err.Error())
// 		}
// 	}()

// }

// func (s *Stack) GetContainerEndpoint(containerName string) string {

// 	switch containerName {
// 	case "nozl":
// 		ep, err := s.nozl.Endpoint(s.ctx, "")

// 		if err != nil {
// 			Logger.Error(err.Error())
// 			return ""
// 		}

// 		return "http://" + ep

// 	case "nats":
// 		ep, err := s.nats.Endpoint(s.ctx, "")

// 		if err != nil {
// 			Logger.Error(err.Error())
// 			return ""
// 		}

// 		return "nats://" + ep

// 	default:
// 		return ""
// 	}
// }

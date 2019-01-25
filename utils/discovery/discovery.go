package discovery

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/lworkltd/kits/helper/consul"
	"github.com/lworkltd/kits/service/discovery"
	"github.com/lworkltd/kits/service/profile"
)

// RegisterServerWithProfile Register the service with profile
//
// `checkUrl` could be either an URL or a path, if ""==checkUrl health check will use tcp
//
// Health check interval use the default value which should be 60s,Health check timeout also
// use the default value which should be 15s
func RegisterServerWithProfile(checkUrl string, cfg *profile.Service) error {
	if !cfg.Reportable {
		return nil
	}

	// Check the profile arguments valid.
	port, err := checkAndResolveProfile(cfg)
	if err != nil {
		return err
	}

	// To warn that service node have move from host to another
	endpoints, ids, err := discovery.Discover(cfg.ReportName)
	if err == nil {
		newEp := net.JoinHostPort(cfg.ReportIp, strconv.Itoa(port))
		for index, ep := range endpoints {
			if cfg.ReportId == ids[index] {
				if ep != newEp {
					// TODO:xxx
				}
			}
			// TODO:Warn which the same ip port belong to differet service
		}
	}

	if "" != checkUrl {
		checkUrl = makeCheckUrl(cfg.ReportIp, port, checkUrl)
	}
	return discovery.Register(&consul.RegisterOption{
		Ip:            cfg.ReportIp,
		Port:          port,
		CheckUrl:      checkUrl,
		Name:          cfg.ReportName,
		Id:            cfg.ReportId,
		Tags:          cfg.ReportTags,
		CheckInterval: cfg.CheckInterval,
		CheckTimeout:  cfg.CheckTimeout,
	})
}

// makeCheckUrl return the health check URL
// If the path is already a complete URL, it do nothing
// If the path is just a route path, return the URL making up whth given ip&port
func makeCheckUrl(ip string, port int, path string) string {
	if strings.HasPrefix(path, "/") {
		return fmt.Sprintf("http://%s:%d%s", ip, port, path)
	}

	return path
}

// RegisterGrpcServerWithProfile register a grpc server to discovery server/agent
// Note 1: health server must be serve on GRPC server,code it like:
// {
// 	server := grpc.NewServer()
// 	// Health server serve
// 	healthServer := health.NewServer()
// 	hv1.RegisterHealthServer(server, healthServer)
// 	// Logic server serve
// 	grpccomm.RegisterCommServiceServer(server, &myServer{})
// 	// Mark the status of logic server
// 	healthServer.SetServingStatus("grpccomm.CommService", hv1.HealthCheckResponse_SERVING)
// }
//
// serverOfHealthCheck could be omitted,which mean the check will always success except any grpc errors
// or one can specify serverOfHealthCheck as name of any served servers.likely,grpccomm.CommService,pb.MyServer,etc.
//
// Note 2: health-check will only work on a consul server/agent with the version bigger than v1.0.6
func RegisterGrpcServerWithProfile(serverOfHealthCheck string, cfg *profile.Service) error {
	if !cfg.Reportable {
		return nil
	}

	// Check the profile arguments valid.
	port, err := checkAndResolveProfile(cfg)
	if err != nil {
		return err
	}

	// Health format is `server[/serverOfHealthCheck]`
	checkUrl := fmt.Sprintf("%s:%d/%s", cfg.ReportIp, port, serverOfHealthCheck)

	return discovery.Register(&consul.RegisterOption{
		ServerType:    consul.ServerTypeGrpc,
		Ip:            cfg.ReportIp,
		Port:          port,
		CheckUrl:      checkUrl,
		Name:          cfg.ReportName,
		Id:            cfg.ReportId,
		Tags:          cfg.ReportTags,
		CheckInterval: cfg.CheckInterval,
		CheckTimeout:  cfg.CheckTimeout,
	})
}

// checkProfile parse the listen port and check if the profile is configured correctly
func checkAndResolveProfile(cfg *profile.Service) (int, error) {
	if cfg.ReportName == "" {
		return 0, fmt.Errorf("register server need a service name")
	}

	if cfg.ReportId == "" {
		return 0, fmt.Errorf("register server need a service id")
	}

	if cfg.ReportIp == "" {
		return 0, fmt.Errorf("register server need a ip address")
	}

	if cfg.ReportIp == "localhost" || strings.HasPrefix(cfg.ReportIp, "127.0.0") {
		return 0, fmt.Errorf("register server ip can not be a loopback address")
	}

	if cfg.ReportPort > 0 {
		return int(cfg.ReportPort), nil
	}

	_, portStr, err := net.SplitHostPort(cfg.Host)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve host port,%s", cfg.Host)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("service port must be a number:%v", port)
	}

	return port, nil
}

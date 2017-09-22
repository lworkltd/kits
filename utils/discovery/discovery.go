package discovery

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/helper/consul"
	"github.com/lworkltd/kits/service/discovery"
	"github.com/lworkltd/kits/service/profile"
)

// RegisterServerWithProfile Register the service with profile
//
// `checkUrl` could be either an URL or a path
//
// Health check interval use the default value which should be 60s,Health check timeout also
// use the default value which should be 15s
func RegisterServerWithProfile(checkUrl string, cfg *profile.Service) error {
	if !cfg.Reportable {
		return nil
	}

	if cfg.ReportName == "" {
		return fmt.Errorf("register server need a service name")
	}

	if cfg.ReportId == "" {
		return fmt.Errorf("register server need a service id")
	}

	if cfg.ReportIp == "" {
		return fmt.Errorf("register server need a ip address")
	}

	if cfg.ReportIp == "localhost" || strings.HasPrefix(cfg.ReportIp, "127.0.0") {
		return fmt.Errorf("register server ip can not be a loopback address")
	}

	if checkUrl == "" {
		return fmt.Errorf("register server need a health check url")
	}

	_, portStr, err := net.SplitHostPort(cfg.Host)
	if err != nil {
		return fmt.Errorf("cannot resolve host port,%s", cfg.Host)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("service port must be a number:%v", port)
	}

	endpoints, ids, err := discovery.Discover(cfg.ReportName)
	if err == nil {
		newEp := net.JoinHostPort(cfg.ReportIp, strconv.Itoa(port))
		for index, ep := range endpoints {
			if cfg.ReportId == ids[index] {
				if ep != newEp {
					logrus.WithFields(logrus.Fields{
						"id":  cfg.ReportId,
						"old": ep,
						"new": newEp,
					}).Warn("Service has same id,but endpoint changed")
				}
			}
			// TODO:Warn which the same ip port belong to differet service
		}
	}

	logrus.WithFields(logrus.Fields{
		"report_name": cfg.ReportName,
		"report_id":   cfg.ReportId,
		"report_ip":   cfg.ReportIp,
		"report_port": port,
		"check_url":   checkUrl,
		"report_tags": cfg.ReportTags,
	}).Info("Register service info")

	return discovery.Register(&consul.RegisterOption{
		Ip:       cfg.ReportIp,
		Port:     port,
		CheckUrl: makeCheckUrl(cfg.ReportIp, port, checkUrl),
		Name:     cfg.ReportName,
		Id:       cfg.ReportId,
		Tags:     cfg.ReportTags,
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

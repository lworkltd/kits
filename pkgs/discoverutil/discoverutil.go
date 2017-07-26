package discoverutil

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/lvhuat/kits/helper/consul"
	"github.com/lvhuat/kits/service/discover"
	"github.com/lvhuat/kits/service/profile"
	"net"
	"strconv"
	"strings"
)

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

	endpoints, ids, err := discover.Discover(cfg.ReportName)
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

	return discover.Register(&consul.RegisterOption{
		Ip:       cfg.ReportIp,
		Port:     port,
		CheckUrl: checkUrl,
		Name:     cfg.ReportName,
		Id:       cfg.ReportId,
		Tags:     cfg.ReportTags,
	})
}

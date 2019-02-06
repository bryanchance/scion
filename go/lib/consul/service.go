// Copyright 2018 Anapaya Systems

package consul

import (
	"context"
	"fmt"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/consul/consulconfig"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/periodic"
)

type Status string

const (
	StatusPass Status = consulapi.HealthPassing
	StatusWarn Status = consulapi.HealthWarning
	StatusCrit Status = consulapi.HealthCritical
)

type CheckInfo struct {
	Status Status
	Note   string
}

type SvcType string

const (
	BS  SvcType = "BeaconService"
	CS  SvcType = "CertificateService"
	DS  SvcType = "DiscoveryService"
	PS  SvcType = "PathService"
	SIG SvcType = "SIG"
)

type Service struct {
	Agent  *consulapi.Agent
	ID     string
	Prefix string
	Addr   *addr.AppAddr
	Type   SvcType
}

// Register attempts to register the service with the consul agent for the connection period.
func (s *Service) Register(connPeriod time.Duration, check func() CheckInfo,
	checkParams consulconfig.HealthCheck) (*periodic.Runner, error) {

	if checkParams.Interval.Duration+checkParams.Timeout.Duration >= checkParams.TTL.Duration {
		return nil, common.NewBasicError("Check interval plus timeout must not exceed TTL", nil,
			"ttl", checkParams.TTL, "interval", checkParams.Interval,
			"timeout", checkParams.Timeout)
	}
	checkName := s.checkName(checkParams.Name)
	info := &consulapi.AgentServiceRegistration{
		ID:      s.ID,
		Name:    s.Name(),
		Address: s.Addr.L3.IP().String(),
		Port:    int(s.Addr.L4.Port()),
		Check: &consulapi.AgentServiceCheck{
			Name:    checkName,
			CheckID: checkName,
			TTL:     checkParams.TTL.String(),
			DeregisterCriticalServiceAfter: checkParams.DeregisterCriticalServiceAfter.String(),
		},
	}
	// Retry registering every second for connPeriod
	var err error
	ticker := time.NewTicker(time.Second)
	timer := time.NewTimer(connPeriod)
	defer ticker.Stop()
	defer timer.Stop()
Top:
	for {
		if err = s.Agent.ServiceRegister(info); err == nil {
			break
		}
		select {
		case <-ticker.C:
		case <-timer.C:
			break Top
		}
	}
	if err != nil {
		return nil, common.NewBasicError("Unable to register service", err)
	}
	t := &TTLSetter{
		Agent: s.Agent,
		ID:    info.Check.CheckID,
		Check: check,
	}
	r := periodic.StartPeriodicTask(t, periodic.NewTicker(checkParams.Interval.Duration),
		checkParams.Timeout.Duration)
	return r, nil
}

func (s *Service) checkName(name string) string {
	if name != "" {
		return name
	}
	return fmt.Sprintf("Health Check: %s", s.ID)
}

// Deregister deregisters the service from the consul agent.
func (s *Service) Deregister() error {
	return s.Agent.ServiceDeregister(s.ID)
}

// Name concatenates the service name with the prefix.
func (s *Service) Name() string {
	if s.Prefix == "" {
		return string(s.Type)
	}
	return fmt.Sprintf("%s%s", s.Prefix, s.Type)
}

var _ (periodic.Task) = (*TTLSetter)(nil)

type TTLSetter struct {
	Agent *consulapi.Agent
	ID    string
	Check func() CheckInfo
}

func (t *TTLSetter) Run(ctx context.Context) {
	info := t.Check()
	err := t.Agent.UpdateTTL(t.ID, info.Note, string(info.Status))
	if err != nil {
		log.Error("[consul] Unable to set status", "err", err)
	}
}

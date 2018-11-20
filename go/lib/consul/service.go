// Copyright 2018 Anapaya Systems

package consul

import (
	"context"
	"time"

	consulapi "github.com/hashicorp/consul/api"

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
	Id     string
	Status Status
	Note   string
}

type SvcType string

const (
	BS SvcType = "BeaconService"
	CS SvcType = "CertificateService"
	DS SvcType = "DiscoveryService"
	PS SvcType = "PathService"
)

type Service struct {
	Type        SvcType
	Prefix      string
	Agent       *consulapi.Agent
	Check       func() CheckInfo
	CheckParams consulconfig.HealthCheck
	Logger      log.Logger
}

func (s *Service) StartUpdateTTL(connPeriod time.Duration) (*periodic.Runner, error) {
	var err error
	t, err := s.InitialConnect(connPeriod)
	if err != nil {
		return nil, err
	}
	r := periodic.StartPeriodicTask(t, periodic.NewTicker(s.CheckParams.Interval.Duration),
		s.CheckParams.Timeout.Duration)
	return r, nil
}

func (s *Service) InitialConnect(connPeriod time.Duration) (*TTLSetter, error) {
	t := &TTLSetter{
		Agent:  s.Agent,
		Check:  s.Check,
		Logger: s.Logger,
	}
	var err error
	ticker := time.NewTicker(time.Second)
	timer := time.NewTimer(connPeriod)
	defer ticker.Stop()
	defer timer.Stop()
Top:
	for {
		if err = t.runTimeout(time.Second); err == nil {
			break
		}
		select {
		case <-ticker.C:
		case <-timer.C:
			break Top
		}
	}
	return t, err
}

var _ (periodic.Task) = (*TTLSetter)(nil)

type TTLSetter struct {
	Agent  *consulapi.Agent
	Check  func() CheckInfo
	Logger log.Logger
}

func (t *TTLSetter) Run(ctx context.Context) {
	if err := t.run(ctx); err != nil && t.Logger != nil {
		t.Logger.Error("Unable to set status", "err", err)
	}
}

func (t *TTLSetter) run(ctx context.Context) error {
	info := t.Check()
	return t.Agent.UpdateTTL(info.Id, info.Note, string(info.Status))
}

func (t *TTLSetter) runTimeout(timeout time.Duration) error {
	ctx, cancelF := context.WithTimeout(context.Background(), timeout)
	defer cancelF()
	return t.run(ctx)
}

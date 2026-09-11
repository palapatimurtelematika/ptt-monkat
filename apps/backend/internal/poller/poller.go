package poller

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gosnmp/gosnmp"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/config"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/models"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/repository"
)

// Poller is the universal SNMP engine. It knows nothing about vendors or
// device types: every target it polls is a row in snmp_targets.
type Poller struct {
	targets  *repository.TargetRepository
	metrics  *repository.MetricRepository
	interval time.Duration
	timeout  time.Duration
	retries  int
	sem      chan struct{} // bounds concurrent SNMP sockets
	done     chan struct{} // closed when Run returns
}

func New(targets *repository.TargetRepository, metrics *repository.MetricRepository, cfg config.Config) *Poller {
	return &Poller{
		targets:  targets,
		metrics:  metrics,
		interval: cfg.PollInterval,
		timeout:  cfg.SNMPTimeout,
		retries:  cfg.SNMPRetries,
		sem:      make(chan struct{}, cfg.PollConcurrency),
		done:     make(chan struct{}),
	}
}

// Run polls until ctx is cancelled. Blocks; start it with `go`.
func (p *Poller) Run(ctx context.Context) {
	defer close(p.done)

	go p.drainWriteErrors(ctx)

	log.Printf("poller: every %s, up to %d concurrent, snmp timeout %s",
		p.interval, cap(p.sem), p.timeout)

	// Poll immediately — otherwise a fresh dashboard shows nothing for a
	// whole interval.
	p.tick(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.metrics.Flush()
			log.Println("poller: stopped, write buffer flushed")
			return
		case <-ticker.C:
			p.tick(ctx)
		}
	}
}

// Wait blocks until Run has returned and flushed, or until timeout.
//
// Cancelling the context is not enough on its own: the caller closes the
// Influx client next, and closing it while Run is still flushing panics with
// "send on closed channel".
func (p *Poller) Wait(timeout time.Duration) {
	select {
	case <-p.done:
	case <-time.After(timeout):
		log.Println("poller: did not stop within timeout")
	}
}

// tick polls every active target once, concurrently.
//
// ponytail: snmp_targets.poll_interval_sec is not honored — every active
// target is polled on every tick. Add a lastPolled map here if devices ever
// need different cadences.
func (p *Poller) tick(ctx context.Context) {
	started := time.Now()

	targets, err := p.targets.FindActive(ctx)
	if err != nil {
		log.Printf("poller: load targets: %v", err)
		return
	}
	if len(targets) == 0 {
		return
	}

	var ok, failed atomic.Int64
	var wg sync.WaitGroup

	for _, t := range targets {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire before touching the network. Unbounded fan-out over a
			// few thousand targets would open that many UDP sockets at once.
			select {
			case p.sem <- struct{}{}:
				defer func() { <-p.sem }()
			case <-ctx.Done():
				return
			}

			if err := p.pollOne(t); err != nil {
				// One dead device must never stop the others.
				log.Printf("poller: %s (%s) %s: %v", t.DeviceName, t.IPAddress, t.MetricName, err)
				failed.Add(1)
				return
			}
			ok.Add(1)
		}()
	}

	wg.Wait()
	log.Printf("poll tick: %d ok, %d failed, took %s", ok.Load(), failed.Load(), time.Since(started).Round(time.Millisecond))
}

func (p *Poller) pollOne(t models.SnmpTarget) error {
	client, err := snmpClient(t, p.timeout, p.retries)
	if err != nil {
		return err
	}
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Conn.Close()

	result, err := client.Get([]string{t.OID})
	if err != nil {
		return fmt.Errorf("get %s: %w", t.OID, err)
	}
	if len(result.Variables) == 0 {
		return fmt.Errorf("get %s: empty response", t.OID)
	}

	raw, err := toFloat(result.Variables[0])
	if err != nil {
		return fmt.Errorf("oid %s: %w", t.OID, err)
	}

	// The multiplier is the calibration knob: raw SNMP counts become dBm,
	// tenths of a degree, whatever the device's MIB scales to.
	multiplier := t.Multiplier
	if multiplier == 0 {
		// Never an intentional calibration — zeroing every reading from a
		// device is worse than ignoring one bad row.
		log.Printf("poller: target %d has multiplier 0, treating as 1", t.ID)
		multiplier = 1
	}

	p.metrics.Write(models.MetricPoint{
		TargetID:   t.ID,
		DeviceName: t.DeviceName,
		IPAddress:  t.IPAddress,
		MetricName: t.MetricName,
		Unit:       t.Unit,
		Value:      raw * multiplier,
		Time:       time.Now(),
	})

	return nil
}

// snmpClient builds a session from the target row alone.
func snmpClient(t models.SnmpTarget, timeout time.Duration, retries int) (*gosnmp.GoSNMP, error) {
	var version gosnmp.SnmpVersion
	switch t.SNMPVersion {
	case "1":
		version = gosnmp.Version1
	case "2c", "":
		version = gosnmp.Version2c
	case "3":
		return nil, fmt.Errorf("snmpv3 needs user/auth/priv credentials that snmp_targets does not carry")
	default:
		return nil, fmt.Errorf("unknown snmp version %q", t.SNMPVersion)
	}

	port := t.SNMPPort
	if port == 0 {
		port = 161
	}
	community := t.SNMPCommunity
	if community == "" {
		community = "public"
	}

	return &gosnmp.GoSNMP{
		Target:    t.IPAddress,
		Port:      port,
		Community: community,
		Version:   version,
		Timeout:   timeout,
		Retries:   retries,
	}, nil
}

// drainWriteErrors consumes the async write client's error channel. Leaving
// it undrained makes failed writes silent.
func (p *Poller) drainWriteErrors(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case err, open := <-p.metrics.Errors():
			if !open {
				return
			}
			log.Printf("poller: influx write: %v", err)
		}
	}
}

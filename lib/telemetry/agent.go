// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

import (
	"context"
	"fmt"
	"log"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/mlog"
)

const (
	defQueueSize = 512
)

// Agent is the one that responsible to collect and forward the metrics.
type Agent struct {
	bulkq   chan []Metric
	singleq chan Metric

	running chan bool

	opts AgentOptions
}

// NewAgent create, initalize, and run the new Agent.
// The agent will start auto collecting the metrics in the background every
// [AgentOptions.Interval] and forward it to each [Forwarder].
func NewAgent(opts AgentOptions) (agent *Agent) {
	opts.init()

	agent = &Agent{
		opts: opts,

		singleq: make(chan Metric, defQueueSize),
		bulkq:   make(chan []Metric, defQueueSize),
		running: make(chan bool, 1),
	}

	go agent.forwarder()
	go agent.collector()

	return agent
}

// BulkForward push list of Metric asynchronously.
// If ctx contains ContextForwardWait, it will forward the metric
// synchronously.
func (agent *Agent) BulkForward(ctx context.Context, list []Metric) error {
	if len(list) == 0 {
		return nil
	}
	var ctxWait = ctx.Value(agentContextForwardWait)
	if ctxWait == nil {
		agent.bulkq <- list
		return nil
	}
	return agent.forwardBulk(ctx, list)
}

func (agent *Agent) collect() (all []Metric) {
	var (
		ts = agent.opts.Timestamp()

		col  Collector
		list []Metric
	)

	for _, col = range agent.opts.Collectors {
		list = col.Collect(ts)
		all = append(all, list...)
	}
	return all
}

// collector collect the metrics on each interval and forward it.
func (agent *Agent) collector() {
	var (
		logp    = `collector`
		ticker  = time.NewTicker(agent.opts.Interval)
		metrics []Metric
		err     error
	)

	for {
		select {
		case <-ticker.C:
			metrics = agent.collect()
			err = agent.BulkForward(context.Background(), metrics)
			if err != nil {
				mlog.Errf(`%s: %s`, logp, err)
			}

		case <-agent.running:
			ticker.Stop()
			// ACK the Stop.
			agent.running <- false
			return
		}
	}
}

// Forward single metric to agent asynchronously.
// If ctx contains ContextForwardWait, it will forward the metric
// synchronously.
func (agent *Agent) Forward(ctx context.Context, m Metric) (err error) {
	var ctxv = ctx.Value(agentContextForwardWait)
	if ctxv == nil {
		agent.singleq <- m
		return nil
	}
	return agent.forwardSingle(ctx, &m)
}

func (agent *Agent) forwardBulk(ctx context.Context, list []Metric) (err error) {
	if len(list) == 0 {
		return nil
	}

	var (
		ts = agent.opts.Timestamp()
		x  int
	)
	for ; x < len(list); x++ {
		if list[x].Timestamp <= 0 {
			list[x].Timestamp = ts
		}
	}

	var (
		// Map of Formatter.Name with its format result.
		fmtWire = map[string][]byte{}

		fwd     Forwarder
		fmter   Formatter
		fmtName string
		wire    []byte
		errfwd  error
		ok      bool
	)
	for _, fwd = range agent.opts.Forwarders {
		fmter = fwd.Formatter()
		if fmter == nil {
			continue
		}

		fmtName = fmter.Name()

		// Check if we have format the metrics before using the same
		// Formatter.
		wire, ok = fmtWire[fmtName]
		if !ok {
			wire = fmter.BulkFormat(list, agent.opts.Metadata)
			fmtWire[fmtName] = wire
		}

		select {
		case <-ctx.Done():
			return err
		default:
			_, errfwd = fwd.Write(wire)
			if errfwd != nil {
				if err == nil {
					err = fmt.Errorf(`forwardBulk: %w`, errfwd)
				} else {
					err = fmt.Errorf(`%w: %w`, err, errfwd)
				}
			}
		}
	}
	return err
}

func (agent *Agent) forwardSingle(ctx context.Context, m *Metric) (err error) {
	if m == nil {
		return nil
	}

	var (
		// Map of Formatter.Name with its format result.
		fmtWire = map[string][]byte{}

		fwd     Forwarder
		fmter   Formatter
		fmtName string
		wire    []byte
		errfwd  error
		ok      bool
	)

	if m.Timestamp <= 0 {
		m.Timestamp = agent.opts.Timestamp()
	}

	for _, fwd = range agent.opts.Forwarders {
		fmter = fwd.Formatter()
		if fmter == nil {
			continue
		}

		fmtName = fmter.Name()

		// Check if we have format the metrics before using the same
		// Formatter.
		wire, ok = fmtWire[fmtName]
		if !ok {
			wire = fmter.Format(*m, agent.opts.Metadata)
			fmtWire[fmtName] = wire
		}

		select {
		case <-ctx.Done():
			// Request cancelled, timeout, or reach deadlines.
			return err
		default:
			_, errfwd = fwd.Write(wire)
			if errfwd != nil {
				if err == nil {
					err = fmt.Errorf(`forwardSingle: %w`, errfwd)
				} else {
					err = fmt.Errorf(`%w: %w`, err, errfwd)
				}
			}
		}
	}
	return err
}

// forwarder the goroutine that queue and forward single or bulk of Metric.
func (agent *Agent) forwarder() {
	var (
		m    Metric
		list []Metric
		err  error
	)

	for {
		select {
		case list = <-agent.bulkq:
			err = agent.forwardBulk(context.Background(), list)
			if err != nil {
				log.Print(err)
			}
		case m = <-agent.singleq:
			err = agent.forwardSingle(context.Background(), &m)
			if err != nil {
				log.Print(err)
			}
		case <-agent.running:
			// ACK the Stop.
			agent.running <- false
			return
		}
	}
}

// Stop the agent and close all [Forwarder].
func (agent *Agent) Stop() {
	// Stop the the first goroutine.
	agent.running <- false
	<-agent.running

	// Stop the the second goroutine.
	agent.running <- false
	<-agent.running

	// Close all forwarders.
	var fwd Forwarder
	for _, fwd = range agent.opts.Forwarders {
		_ = fwd.Close()
	}
}

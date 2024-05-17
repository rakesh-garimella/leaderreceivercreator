// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package leaderelectionreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"github.com/skhalash/leaderelectionreceiver/internal/sharedcomponent"
	"github.com/skhalash/leaderelectionreceiver/internal/metadata"
)

// This file implements factory for receiver_creator. A receiver_creator can create other receivers at runtime.

var receivers = sharedcomponent.NewSharedComponents()

// NewFactory creates a factory for receiver creator.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithLogs(createLogsReceiver, metadata.LogsStability),
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability),
		receiver.WithTraces(createTracesReceiver, metadata.TracesStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		receiverTemplates: map[string]receiverTemplate{},
	}
}

func createLogsReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	r := receivers.GetOrAdd(cfg, func() component.Component {
		return newReceiverCreator(params, cfg.(*Config))
	})
	r.Component.(*receiverCreator).nextLogsConsumer = consumer
	return r, nil
}

func createMetricsReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	r := receivers.GetOrAdd(cfg, func() component.Component {
		return newReceiverCreator(params, cfg.(*Config))
	})
	r.Component.(*receiverCreator).nextMetricsConsumer = consumer
	return r, nil
}

func createTracesReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Traces,
) (receiver.Traces, error) {
	r := receivers.GetOrAdd(cfg, func() component.Component {
		return newReceiverCreator(params, cfg.(*Config))
	})
	r.Component.(*receiverCreator).nextTracesConsumer = consumer
	return r, nil
}

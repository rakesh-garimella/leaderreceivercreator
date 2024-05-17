// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package receivercreator // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator"

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

var _ receiver.Metrics = (*receiverCreator)(nil)

// receiverCreator implements consumer.Metrics.
type receiverCreator struct {
	params              receiver.CreateSettings
	cfg                 *Config
	nextLogsConsumer    consumer.Logs
	nextMetricsConsumer consumer.Metrics
	nextTracesConsumer  consumer.Traces

	host   component.Host
	cancel context.CancelFunc
}

func newReceiverCreator(params receiver.CreateSettings, cfg *Config) receiver.Metrics {
	return &receiverCreator{
		params: params,
		cfg:    cfg,
	}
}

// Start receiver_creator.
func (rc *receiverCreator) Start(ctx context.Context, host component.Host) error {
	rc.host = host
	ctx = context.Background()
	ctx, rc.cancel = context.WithCancel(ctx)

	go func() {
		for _, template := range rc.cfg.receiverTemplates {
			rc.params.TelemetrySettings.Logger.Info("starting receiver",
				zap.String("name", template.id.String()))

			consumer, err := newEnhancingConsumer(rc.nextLogsConsumer, rc.nextMetricsConsumer, rc.nextTracesConsumer)
			if err != nil {
				rc.params.TelemetrySettings.Logger.Error("failed creating resource enhancer", zap.String("receiver", template.id.String()), zap.Error(err))
				continue
			}

			runner := newReceiverRunner(rc.params, rc.host)
			_, err = runner.start(
				receiverConfig{
					id:     template.id,
					config: template.config,
				},
				consumer,
			)
			if err != nil {
				rc.params.TelemetrySettings.Logger.Error("failed to start receiver", zap.String("receiver", template.id.String()), zap.Error(err))
				continue
			}
		}
	}()

	return nil
}

// Shutdown stops the receiver_creator and all its receivers started at runtime.
func (rc *receiverCreator) Shutdown(context.Context) error {
	rc.cancel()
	return nil
}

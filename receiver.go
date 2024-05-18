// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package leaderelectionreceiver

import (
	"fmt"
	"context"
	"os"
	"path/filepath"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var _ receiver.Metrics = (*leaderElectionReceiver)(nil)

// leaderElectionReceiver implements consumer.Metrics.
type leaderElectionReceiver struct {
	params              receiver.CreateSettings
	cfg                 *Config
	nextLogsConsumer    consumer.Logs
	nextMetricsConsumer consumer.Metrics
	nextTracesConsumer  consumer.Traces

	host   component.Host
	cancel context.CancelFunc
}

func newReceiverCreator(params receiver.CreateSettings, cfg *Config) receiver.Metrics {
	return &leaderElectionReceiver{
		params: params,
		cfg:    cfg,
	}
}

// Start receiver_creator.
func (rc *leaderElectionReceiver) Start(ctx context.Context, host component.Host) error {
	rc.host = host
	ctx = context.Background()
	ctx, rc.cancel = context.WithCancel(ctx)

	rc.params.TelemetrySettings.Logger.Info("Starting leader election receiver...")

	client, err := rc.newClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	rc.params.TelemetrySettings.Logger.Info("Creating leader elector...")

	leaderElector, err := newLeaderElector(
		client,
		func(ctx context.Context) {
			rc.params.TelemetrySettings.Logger.Info("Elected as leader")
		},
		func() {
			rc.params.TelemetrySettings.Logger.Info("Lost leadership")
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create leader elector: %w", err)
	}

	leaderElector.Run(ctx)
	return nil
}

func (rc *leaderElectionReceiver) newClient() (kubernetes.Interface, error) {
	kubeConfigPath := filepath.Join(os.Getenv("HOME"), ".kube/config")

	config, err := rest.InClusterConfig()
	if err != nil {
		rc.params.TelemetrySettings.Logger.Warn("Cannot find in cluster config", zap.Error(err))
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			rc.params.TelemetrySettings.Logger.Error("Cannot build ClientConfig", zap.Error(err))
			return nil, err
		}
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		rc.params.TelemetrySettings.Logger.Error("Cannot create Kubernetes client", zap.Error(err))
		return nil, err
	}
	return client, nil
}

func (rc *leaderElectionReceiver) startReceiverRunner() error {
	for _, template := range rc.cfg.receiverTemplates {
		rc.params.TelemetrySettings.Logger.Info("starting receiver",
			zap.String("name", template.id.String()))

		runner := newReceiverRunner(rc.params, rc.host)
		_, err := runner.start(
			receiverConfig{
				id:     template.id,
				config: template.config,
			},
			rc.nextLogsConsumer,
			rc.nextMetricsConsumer,
			rc.nextTracesConsumer,
		)
		if err != nil {
			return fmt.Errorf("failed to start receiver %s: %w", template.id.String(), err)
		}
	}
	return nil
}

// Shutdown stops the receiver_creator and all its receivers started at runtime.
func (rc *leaderElectionReceiver) Shutdown(context.Context) error {
	rc.cancel()
	return nil
}

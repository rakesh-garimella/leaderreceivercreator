// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package leaderreceivercreator

import (
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

const (
	// receiversConfigKey is the config key name used to specify the subreceivers.
	subreceiverConfigKey    = "receiver"
	leaderElectionConfigKey = "leader_election"
)

type leaderElectionConfig struct {
	leaseName            string
	leaseNamespace       string
	leaseDurationSeconds time.Duration
	renewDeadlineSeconds time.Duration
	retryPeriodSeconds   time.Duration
}

// receiverConfig describes a receiver instance with a default config.
type receiverConfig struct {
	// id is the id of the subreceiver (ie <receiver type>/<id>).
	id component.ID
	// config is the map configured by the user in the config file. It is the contents of the map from
	// the "config" section. The keys and values are arbitrarily configured by the user.
	config map[string]any
}

// and its arbitrary config map values.
func newReceiverConfig(name string, cfg map[string]any) (receiverConfig, error) {
	id := component.ID{}
	if err := id.UnmarshalText([]byte(name)); err != nil {
		return receiverConfig{}, fmt.Errorf("failed to parse subreceiver id %v: %w", name, err)
	}

	return receiverConfig{
		id:     id,
		config: cfg,
	}, nil
}

func newLeaderElectionConfig(cfg map[string]any) leaderElectionConfig {
	return leaderElectionConfig{
		leaseName:            cfg["lease_name"].(string),
		leaseNamespace:       cfg["lease_namespace"].(string),
		leaseDurationSeconds: cfg["lease_duration_seconds"].(time.Duration),
		renewDeadlineSeconds: cfg["renew_deadline_seconds"].(time.Duration),
		retryPeriodSeconds:   cfg["retry_period_seconds"].(time.Duration),
	}
}

var _ confmap.Unmarshaler = (*Config)(nil)

// Config defines configuration for receiver_creator.
type Config struct {
	leaderElectionConfig leaderElectionConfig
	subreceiverConfig    receiverConfig
}

func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	if componentParser == nil {
		// Nothing to do if there is no config given.
		return nil
	}

	if err := componentParser.Unmarshal(cfg, confmap.WithIgnoreUnused()); err != nil {
		return err
	}

	subreceiverConfig, err := componentParser.Sub(subreceiverConfigKey)
	if err != nil {
		return fmt.Errorf("unable to extract key %v: %w", subreceiverConfigKey, err)
	}

	leaderElectionConfig, err := componentParser.Sub(leaderElectionConfigKey)
	if err != nil {
		return fmt.Errorf("unable to extract key %v: %w", leaderElectionConfigKey, err)
	}

	for subreceiverKey := range subreceiverConfig.ToStringMap() {
		receiverConfig, err := subreceiverConfig.Sub(subreceiverKey)
		if err != nil {
			return fmt.Errorf("unable to extract subreceiver key %v: %w", subreceiverKey, err)
		}

		cfg.subreceiverConfig, err = newReceiverConfig(subreceiverKey, receiverConfig.ToStringMap())
		if err != nil {
			return fmt.Errorf("failed to create subreceiver config: %w", err)
		}

		cfg.leaderElectionConfig = newLeaderElectionConfig(leaderElectionConfig.ToStringMap())

		return nil
	}

	return nil
}

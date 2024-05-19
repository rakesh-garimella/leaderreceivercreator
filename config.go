// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package leaderelectionreceiver

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

const (
	// receiversConfigKey is the config key name used to specify the subreceivers.
	subreceiverConfigKey = "receiver"
)

// receiverConfig describes a receiver instance with a default config.
type receiverConfig struct {
	// id is the id of the subreceiver (ie <receiver type>/<id>).
	id component.ID
	// config is the map configured by the user in the config file. It is the contents of the map from
	// the "config" section. The keys and values are arbitrarily configured by the user.
	config userConfigMap
}

// userConfigMap is an arbitrary map of string keys to arbitrary values as specified by the user
type userConfigMap map[string]any

// and its arbitrary config map values.
func newReceiverConfig(name string, cfg userConfigMap) (receiverConfig, error) {
	id := component.ID{}
	if err := id.UnmarshalText([]byte(name)); err != nil {
		return receiverConfig{}, fmt.Errorf("failed to parse subreceiver id %v: %w", name, err)
	}

	return receiverConfig{
		id:     id,
		config: cfg,
	}, nil
}

var _ confmap.Unmarshaler = (*Config)(nil)

// Config defines configuration for receiver_creator.
type Config struct {
	subreceiverConfig receiverConfig
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

	for subreceiverKey := range subreceiverConfig.ToStringMap() {
		receiverConfig, err := subreceiverConfig.Sub(subreceiverKey)
		if err != nil {
			return fmt.Errorf("unable to extract subreceiver key %v: %w", subreceiverKey, err)
		}

		cfg.subreceiverConfig, err = newReceiverConfig(subreceiverKey, receiverConfig.ToStringMap())
		if err != nil {
			return fmt.Errorf("failed to create subreceiver config: %w", err)
		}

		return nil
	}

	return nil
}

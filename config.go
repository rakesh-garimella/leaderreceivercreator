// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package receivercreator // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator"

import (
	"fmt"

	"github.com/spf13/cast"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
)

const (
	// receiversConfigKey is the config key name used to specify the subreceivers.
	receiversConfigKey = "receivers"
	// endpointConfigKey is the key name mapping to ReceiverSettings.Endpoint.
	endpointConfigKey = "endpoint"
	// configKey is the key name in a subreceiver.
	configKey = "config"
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

// receiverTemplate is the configuration of a single subreceiver.
type receiverTemplate struct {
	receiverConfig
}

// resourceAttributes holds a map of default resource attributes for each Endpoint type.
type resourceAttributes map[observer.EndpointType]map[string]string

// newReceiverTemplate creates a receiverTemplate instance from the full name of a subreceiver
// and its arbitrary config map values.
func newReceiverTemplate(name string, cfg userConfigMap) (receiverTemplate, error) {
	id := component.ID{}
	if err := id.UnmarshalText([]byte(name)); err != nil {
		return receiverTemplate{}, err
	}

	return receiverTemplate{
		receiverConfig: receiverConfig{
			id:     id,
			config: cfg,
		},
	}, nil
}

var _ confmap.Unmarshaler = (*Config)(nil)

// Config defines configuration for receiver_creator.
type Config struct {
	receiverTemplates map[string]receiverTemplate
	// WatchObservers are the extensions to listen to endpoints from.
	WatchObservers []component.ID `mapstructure:"watch_observers"`
}

func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	if componentParser == nil {
		// Nothing to do if there is no config given.
		return nil
	}

	if err := componentParser.Unmarshal(cfg, confmap.WithIgnoreUnused()); err != nil {
		return err
	}

	receiversCfg, err := componentParser.Sub(receiversConfigKey)
	if err != nil {
		return fmt.Errorf("unable to extract key %v: %w", receiversConfigKey, err)
	}

	for subreceiverKey := range receiversCfg.ToStringMap() {
		subreceiverSection, err := receiversCfg.Sub(subreceiverKey)
		if err != nil {
			return fmt.Errorf("unable to extract subreceiver key %v: %w", subreceiverKey, err)
		}
		cfgSection := cast.ToStringMap(subreceiverSection.Get(configKey))
		subreceiver, err := newReceiverTemplate(subreceiverKey, cfgSection)
		if err != nil {
			return err
		}

		// Unmarshals receiver_creator configuration like rule.
		if err = subreceiverSection.Unmarshal(&subreceiver, confmap.WithIgnoreUnused()); err != nil {
			return fmt.Errorf("failed to deserialize sub-receiver %q: %w", subreceiverKey, err)
		}

		cfg.receiverTemplates[subreceiverKey] = subreceiver
	}

	return nil
}

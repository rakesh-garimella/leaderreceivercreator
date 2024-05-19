package leaderelectionreceiver

import (
	"testing"
	"path/filepath"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"github.com/stretchr/testify/require"
	"github.com/skhalash/leaderelectionreceiver/internal/metadata"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	tests := []struct {
		id       component.ID
		expected component.Config
	}{
		{
			id: component.NewID(metadata.Type),
			expected: &Config{
				subreceiverConfig: receiverConfig{
					id: component.MustNewID("otlp"),
					config: userConfigMap{
						"protocols": map[string]any{
							"grpc": nil,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, component.UnmarshalConfig(sub, cfg))

			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

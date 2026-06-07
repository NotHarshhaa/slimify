// Package config handles loading and parsing of slimify configuration files.
// It supports slimify.yaml, .slimifyrc, .slimifyrc.json, .slimifyrc.yaml,
// and slimify.config.js via viper (cosmiconfig-compatible).
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the full slimify configuration.
type Config struct {
	Ignore IgnoreConfig `mapstructure:"ignore" yaml:"ignore"`
	Audit  AuditConfig  `mapstructure:"audit" yaml:"audit"`
	Fix    FixConfig    `mapstructure:"fix" yaml:"fix"`
}

// IgnoreConfig controls .dockerignore generation behavior.
type IgnoreConfig struct {
	// Whitelist contains paths that should never be ignored, even if detected as bloat.
	Whitelist []string `mapstructure:"whitelist" yaml:"whitelist"`
	// Blacklist contains paths that should always be ignored, even if not auto-detected.
	Blacklist []string `mapstructure:"blacklist" yaml:"blacklist"`
}

// AuditConfig controls audit behavior.
type AuditConfig struct {
	// ThresholdMB is the minimum file size (in MB) to flag. Default: 1.
	ThresholdMB float64 `mapstructure:"threshold_mb" yaml:"threshold_mb"`
	// TopFilesPerLayer is how many large files to show per layer. Default: 10.
	TopFilesPerLayer int `mapstructure:"top_files_per_layer" yaml:"top_files_per_layer"`
}

// FixConfig controls the fix command behavior.
type FixConfig struct {
	// BaseImage overrides the auto-selected base image.
	BaseImage string `mapstructure:"base_image" yaml:"base_image"`
	// MultiStage forces multi-stage rewrite when true.
	MultiStage bool `mapstructure:"multi_stage" yaml:"multi_stage"`
	// OutputDir is the directory for generated files.
	OutputDir string `mapstructure:"output_dir" yaml:"output_dir"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Audit: AuditConfig{
			ThresholdMB:      1.0,
			TopFilesPerLayer: 10,
		},
		Fix: FixConfig{
			MultiStage: true,
			OutputDir:  ".",
		},
	}
}

// Load reads the configuration from the given path, or auto-discovers it
// from the current directory. It returns the merged config with defaults.
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Search for config in current directory
		cwd, err := os.Getwd()
		if err != nil {
			return cfg, nil
		}

		v.SetConfigName("slimify")
		v.AddConfigPath(cwd)

		// Also support .slimifyrc variants
		for _, name := range []string{".slimifyrc", ".slimifyrc.yaml", ".slimifyrc.json"} {
			path := filepath.Join(cwd, name)
			if _, err := os.Stat(path); err == nil {
				v.SetConfigFile(path)
				break
			}
		}
	}

	if err := v.ReadInConfig(); err != nil {
		// Config file not found is fine — we use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return cfg, nil
		}
		// If a specific config was requested and failed, that's an error
		if configPath != "" {
			return nil, err
		}
		return cfg, nil
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadFromDir loads configuration from a specific directory.
func LoadFromDir(dir string) (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("slimify")
	v.AddConfigPath(dir)

	// Also search for .slimifyrc variants
	for _, name := range []string{".slimifyrc", ".slimifyrc.yaml", ".slimifyrc.json"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			v.SetConfigFile(path)
			break
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return cfg, nil
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

package cmdutil

import (
	"fmt"
	"k8s.io/klog"
	"path/filepath"

	"github.com/spf13/viper"
)

// InitConfig constructs the configuration from a local configuration file
// or environment variables if available. This is placed in the global `viper`
// instance.
func InitConfig(name string) error {
	cfgDir := "/etc/config"
	cfgName := fmt.Sprintf("%s.yaml", name)
	cfgPath := filepath.Join(cfgDir, cfgName)

	viper.SetConfigFile(cfgPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	klog.Infof("Using config file: %s", viper.ConfigFileUsed())
	return nil
}

// ConfigPath returns the directory path being used by config.
func ConfigPath() string {
	return filepath.Dir(viper.ConfigFileUsed())
}

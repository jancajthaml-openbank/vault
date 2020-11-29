package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestGetConfig(t *testing.T) {
	for _, v := range os.Environ() {
		k := strings.Split(v, "=")[0]
		if strings.HasPrefix(k, "VAULT") {
			os.Unsetenv(k)
		}
	}

	t.Log("has defaults for all values")
	{
		config := LoadConfig()

		if config.Tenant != "" {
			t.Errorf("Tenant default value is not empty")
		}
		if config.LakeHostname != "127.0.0.1" {
			t.Errorf("LakeHostname default value is not 127.0.0.1")
		}
		if config.RootStorage != "/data/t_" {
			t.Errorf("RootStorage default value is not /data/t_")
		}
		if config.LogLevel != "INFO" {
			t.Errorf("LogLevel default value is not INFO")
		}
		if config.MetricsContinuous != true {
			t.Errorf("MetricsContinuous default value is not true")
		}
		if config.MetricsRefreshRate != time.Second {
			t.Errorf("MetricsRefreshRate default value is not 1s")
		}
		if config.MetricsOutput != "/tmp/vault-unit-metrics" {
			t.Errorf("MetricsOutput default value is not /tmp/vault-unit-metrics")
		}
		if config.SnapshotSaturationTreshold != 100 {
			t.Errorf("SnapshotSaturationTreshold default value is not 100")
		}
	}
}

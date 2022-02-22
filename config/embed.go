package config

import (
	_ "embed"
	"os"
	"tczbgo/config"
)

var (
	env string
	//go:embed appsettings.dev.json
	appSettingsDEV []byte
	//go:embed appsettings.local.json
	appSettingsLOCAL []byte
	//go:embed appsettings.prd.json
	appSettingsPRD []byte
	//go:embed appsettings.bak.json
	appSettingsBAK []byte
)

func init() {
	if os.Getenv("env") == "" {
		os.Setenv("env", env)
	}
	config.InitConfig(appSettingsDEV, appSettingsLOCAL, appSettingsPRD, appSettingsBAK)
}

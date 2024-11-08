package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	apiToken string
	projects []Project
	logger *zap.SugaredLogger
}

type Project struct {
	ProjectId string
	Severity string
	Resolved bool
	PollingInterval int
}

func newConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	c := &Config{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
			return nil, err
	}

	configPath := homeDir + "/.config/airbrakeNotify"

	logger, err := NewLogger(configPath)
	if err != nil {
		panic(err.Error())
	}
	defer logger.Sync()
	c.logger = logger.Sugar()

	viper.AddConfigPath(configPath)

	if err = viper.ReadInConfig(); err != nil {
		return c, err
	}

	c.apiToken = viper.GetString("apiToken")

	var projects []Project
	err = viper.UnmarshalKey("projects", &projects)
	if err != nil {
		return c, errors.New(fmt.Sprint("malformed projects in config file: ", err.Error()))
	}

	c.projects = projects

	if c.apiToken == "" {
		return c, errors.New("apiToken is missing")
	}

	return c, nil
}

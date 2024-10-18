package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	apiToken string
	projects []Project
}

type Project struct {
	ProjectId string
	Severity string
	PollingInterval int
}

func newConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/home/wolfgang/.config/airbrakeNotify")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	c := &Config{
		apiToken: viper.GetString("apiToken"),
	}

	var projects []Project
	err := viper.UnmarshalKey("projects", &projects)
	if err != nil {
		return nil, errors.New(fmt.Sprint("malformed projects in config file: ", err.Error()))
	}

	c.projects = projects

	if c.apiToken == "" {
		return nil, errors.New("apiToken is missing")
	}

	return c, nil
}

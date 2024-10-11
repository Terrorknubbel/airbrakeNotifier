package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	airbrakeToken string
	projects []Project
}

type Project struct {
	ProjectId string
	Severity string
}

func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/home/wolfgang/.config/airbrakeNotify")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	c := &Config{
		airbrakeToken: viper.GetString("airbrakeToken"),
	}

	var projects []Project
	err := viper.UnmarshalKey("projects", &projects)
	if err != nil {
		return nil, errors.New(fmt.Sprint("malformed projects in config file: ", err.Error()))
	}

	c.projects = projects

	if c.airbrakeToken == "" {
		return nil, errors.New("airbrakeToken is missing")
	}

	if c.projects[0].Severity == "" {
		c.projects[0].Severity = "error"
	}

	return c, nil
}

package config

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var (
	configFile = os.Getenv("CONFIG")
)

// Config the phabrick config
type Config struct {
	Slack       SlackConfig       `yaml:"slack"`
	Channels    ChannelsConfig    `yaml:"channels"`
	Phabricator PhabricatorConfig `yaml:"phabricator"`
}

// SlackConfig the slack config
type SlackConfig struct {
	Token        string `yaml:"token"`
	Username     string `yaml:"username"`
	ShowAssignee bool   `yaml:"showAssignee"`
	ShowAuthor   bool   `yaml:"showAuthor"`
}

// ChannelsConfig the channels config
type ChannelsConfig struct {
	ObjectTypes []string          `yaml:"objectTypes"`
	Projects    map[string]string `yaml:"projects"`
}

// PhabricatorConfig the phabricator config
type PhabricatorConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

// GetConfig get a phabrick config
func GetConfig() *Config {
	var c *Config
	if configFile == "" {
		configFile = "./phabrick.yaml"
	}
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Unable to locate config file, %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Invalid config file format, %v", err)
	}
	if c.Phabricator.URL == "" {
		c.Phabricator.URL = os.Getenv("PHABRICATOR_URL")
	}
	if c.Phabricator.Token == "" {
		c.Phabricator.Token = os.Getenv("PHABRICATOR_TOKEN")
	}
	return c
}

package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Budget2ConfigStruct struct {
	Payday int `yaml:"payday"`
	Rentday int `yaml:"rentday"`
	Rentamount float32 `yaml:"rentamount"`
	InitialValues map[int]float32 `yaml:"initial-values"`
	SenderAddress string
	EmailRecipients []string `yaml:"email-recipients"`
}

var Budget2Config Budget2ConfigStruct
const configFilename string = "./config.yaml"
const configFilenameSecret string = "./config-secret.yaml"
const configSenderAddress string = "budget2@wold.noclab.com.au"

func ReadConfig() error {

	// Read our public configuration:
	yamlFile, err := ioutil.ReadFile(configFilename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &Budget2Config)
	if err != nil {
		return err
	}

	// Read our secret configuration:
	yamlFileSecret, err := ioutil.ReadFile(configFilenameSecret)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFileSecret, &Budget2Config)
	if err != nil {
		return err
	}

	Budget2Config.SenderAddress = configSenderAddress

	return nil
}

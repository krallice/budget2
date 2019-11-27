package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	)

type Budget2ConfigStruct struct {
	Payday int	`yaml:"payday"`
}

var Budget2Config Budget2ConfigStruct
const configFilename string = "./config.yaml"

func ReadConfig() (error) {
	yamlFile, err := ioutil.ReadFile(configFilename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &Budget2Config)
	if err != nil {
		return err
	}
	return nil
}


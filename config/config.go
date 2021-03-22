package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"errors"
	"os"
	"strconv"
	"strings"
	// "fmt"
)

type Budget2ConfigStruct struct {
	Payday int `yaml:"payday"`
	Rentday int `yaml:"rentday"`
	Rentamount float32 `yaml:"rentamount"`
	InitialValues map[int]float32 `yaml:"initial-values"`
	SenderAddress string
	EmailRecipients []string `yaml:"email-recipients"`
	DBUsername string `yaml:"db-username"`
	DBPassword string `yaml:"db-password"`
	DBServer string `yaml:"db-server"`
	DBName string `yaml:"db-name"`
}

var Budget2Config Budget2ConfigStruct
const configFilename string = "./config.yaml"
const configFilenameSecret string = "./config-secret.yaml"
const configSenderAddress string = "budget2@wold.noclab.com.au"

// Detects if the server is being run inside K8s, and parses environment variables accordingly
// Returns an int indicating if K8s was detected, 0=NO, 1=YES, along with error
func detectAndParseK8sEnvVariables() (int, error) {

	if (len(os.Getenv("KUBERNETES_SERVICE_HOST")) > 0) {

		val, ok := os.LookupEnv("B2_PAYDAY")
		if !ok {
			return 1, errors.New("Environment key B2_PAYDAY not found")
		} else {
			i, err := strconv.Atoi(val)
			if err != nil {
				return 1, err
			}
			Budget2Config.Payday = i
		}

		val, ok = os.LookupEnv("B2_RENTDAY")
		if !ok {
			return 1, errors.New("Environment key B2_RENTDAY not found")
		} else {
			i, err := strconv.Atoi(val)
			if err != nil {
				return 1, err
			}
			Budget2Config.Rentday = i
		}

		val, ok = os.LookupEnv("B2_RENTAMOUNT")
		if !ok {
			return 1, errors.New("Environment key B2_RENTAMOUNT not found")
		} else {
			i, err := strconv.ParseFloat(val, 32)
			if err != nil {
				return 1, err
			}
			Budget2Config.Rentamount = float32(i)
		}

		val, ok = os.LookupEnv("B2_INITIALVALUES")
		if !ok {
			return 1, errors.New("Environment key B2_INITIALVALUES not found")
		} else {
			Budget2Config.InitialValues = make(map[int]float32)
			values := strings.Fields(val)
			for _, value := range values {

				kv := strings.Split(value, ":")
				iv_index, err := strconv.Atoi(kv[0])
				if err != nil {
					return 1, err
				}
				iv_value, err := strconv.ParseFloat(kv[1], 32)
				if err != nil {
					return 1, err
				}
				Budget2Config.InitialValues[iv_index] = float32(iv_value)
			}
		}

		val, ok = os.LookupEnv("B2_SENDERADDRESS")
		if !ok {
			return 1, errors.New("Environment key B2_SENDERADDRESS not found")
		} else {
			Budget2Config.SenderAddress = val
		}

		val, ok = os.LookupEnv("B2_EMAILRECIPIENTS")
		if !ok {
			return 1, errors.New("Environment key B2_EMAILRECIPIENTS not found")
		} else {
			recipients := strings.Fields(val)
			Budget2Config.EmailRecipients = recipients
		}

		val, ok = os.LookupEnv("B2_DBUSERNAME")
		if !ok {
			return 1, errors.New("Environment key B2_DBUSERNAME not found")
		} else {
			Budget2Config.DBUsername = val
		}

		val, ok = os.LookupEnv("B2_DBPASSWORD")
		if !ok {
			return 1, errors.New("Environment key B2_DBPASSWORD not found")
		} else {
			Budget2Config.DBPassword = val
		}

		val, ok = os.LookupEnv("B2_DBSERVER")
		if !ok {
			return 1, errors.New("Environment key B2_DBSERVER not found")
		} else {
			Budget2Config.DBServer = val
		}

		val, ok = os.LookupEnv("B2_DBNAME")
		if !ok {
			return 1, errors.New("Environment key B2_DBNAME not found")
		} else {
			Budget2Config.DBName = val
		}

		return 1, nil

	} else {
		return 0, nil
	}
}

func ReadConfig() error {

	k8s, err := detectAndParseK8sEnvVariables()
	if err != nil {
		return err
	}

	// Running natively outside of k8s:
	if k8s == 0 {

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
	}

	// Hardset sender address for now
	Budget2Config.SenderAddress = configSenderAddress
	return nil
}

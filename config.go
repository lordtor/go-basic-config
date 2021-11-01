package base_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"strings"

	"github.com/imdario/mergo"

	"gopkg.in/resty.v1"
	"gopkg.in/yaml.v2"
	"github.com/lordtor/go-logging"
)

type Secrets map[string]string
type Congig interface {
	GetSecretsFromJson() string
	GetParamsFromYml(path string)
	ParseCloudFile()
	ReloadConfig()
	PrintConfigToLog()
	UpdateSecrets()
}
type ApplicationConfig struct {
	AppName       string `yaml:"app_name"`
	ConfServerURI string `yaml:"conf_server_uri"`
	LogLevel      string `yaml:"log_level"`
	ProfileName   string `yaml:"profile_name"`
	Secrets
}

var (
	//AppConfig is ...
	AppConfig = ApplicationConfig{}
	log       = logging.Log
)

func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

//===========read secrets from APP_NAME.json =========
func (conf *ApplicationConfig) GetSecretsFromJson(fileName string) (Secrets, string, error) {
	if fileName == "" {
		fileName = filepath.Base(os.Args[0])
	}

	filename, _ := filepath.Abs(fmt.Sprintf("%s.json", fileName))
	log.Debug("[Config:getSecretsFromJson]:: from", filename)
	S := Secrets{}
	exist, err := Exists(filename)
	if err == nil && exist {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, filename, err
		}
		err = json.Unmarshal(data, &S)
		if err != nil {
			return nil, filename, err
		}
		return S, filename, nil
	} else {
		return nil, filename, err
	}
}

//===========read params from application.yml =========
func (conf *ApplicationConfig) GetParamsFromYml(path string) error {

	if path == "" {
		path, _ = filepath.Abs("./application.yml")
	}
	log.Debug("[Config:GetParamsFromYml]:: Load file: ", path)
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("[Config:GetParamsFromYml]:: cannot open file: ", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Error("[Config:GetParamsFromYml]:: cannot unmarshal data: ", err)
		return err
	}
	log.Info(conf)
	err = mergo.Merge(&ApplicationConfig{}, conf)
	if err != nil {
		log.Error("[Config:GetParamsFromYml]:: cannot Merge data: ", err)
		return err
	}
	return nil
}

//===========read params from env =========
func GetValueByNameFromEnv(aName string) string {
	res := ""
	log.Debug("[Config:GetValueByNameFromEnv]:: Find env value: ", aName)
	aValue, exists := os.LookupEnv(aName)
	if exists {
		log.Debug("[Config:GetValueByNameFromEnv]:: Load env value: ", aName)
		res = aValue
	}
	return res
}

//read config from Spring Cloud Config Server
func FetchFileFromCloud(AppName string, ProfileName string, ConfServerURI string) ([]byte, error) {
	u, err := url.Parse(ConfServerURI)
	if err != nil && ConfServerURI != "" {
		return nil, err
	} else if ConfServerURI == "" {
		return nil, errors.New("ConfServerURI is empty")
	}
	if AppName != "" && ProfileName != "" {
		u.Path = path.Join(u.Path,
			strings.Join([]string{AppName, "-", ProfileName, ".yml"},
				""))
		log.Info("[Config:fetchFileFromCloud]:: Load env value: ", u.Path)
	} else if AppName != "" && ProfileName == "" {
		u.Path = path.Join(u.Path,
			strings.Join([]string{AppName, ".yml"},
				""))
		log.Info("[Config:fetchFileFromCloud]:: Load env value: ", u.Path)
	} else {
		log.Warn("[Config:fetchFileFromCloud]:: Not set : AppName")
		return nil, nil
	}
	link := u.String()

	resp, err := resty.R().Get(link)
	if err != nil {
		log.Error("[Config:fetchFileFromCloud]:: ", err)
		return nil, err
	}
	return resp.Body(), nil
}

func (conf *ApplicationConfig) ParseCloudFile() (*ApplicationConfig, error) {
	config := &ApplicationConfig{}
	rawBytes, err := FetchFileFromCloud(conf.AppName, conf.ProfileName, conf.ConfServerURI)
	if err != nil {
		return nil, err
	}
	log.Debug(string(rawBytes))
	err = yaml.Unmarshal(rawBytes, &config)
	if err != nil {
		log.Error("[Config:parseCloudFile]:: ", err)
		return nil, err
	}
	log.Debug("[Config:parseCloudFile]:: parse cloud file: ", conf)
	return config, nil
}

func (conf *ApplicationConfig) ReloadConfig() {
	log.Debug("[Config:ReloadConfig]:: Start func ReloadConfig")
	err := conf.GetParamsFromYml("")
	if err != nil {
		log.Error(err)
	}

	ConfServerURI := GetValueByNameFromEnv("OMNI_GLOBAL_SPRING_CLOUD_CONFIG_URI")
	AppName := GetValueByNameFromEnv("APP_NAME")
	ProfileName := GetValueByNameFromEnv("PROFILE_NAME")
	if ConfServerURI != "" {
		conf.ConfServerURI = ConfServerURI
	}
	if AppName != "" {
		conf.AppName = AppName
	}
	if ProfileName != "" {
		conf.ProfileName = ProfileName
	}
	log.Debug("[Config:ReloadConfig]:: OMNI_GLOBAL_SPRING_CLOUD_CONFIG_URI", conf.ConfServerURI)
	if ProfileName != "develop" {
		cloud, err := conf.ParseCloudFile()
		if err != nil && conf.ConfServerURI != "" {
			log.Fatal(err)
		}
		if cloud != nil {
			err := mergo.Merge(&conf, cloud)
			if err != nil {
				log.Error(err)
			}
		}
	}
	secrets, file, err := conf.GetSecretsFromJson("")
	if err != nil {
		log.Error("[Config:ReloadConfig]:: ", err)
	} else {
		log.Infof("[Config:ReloadConfig]:: Use credential's from different file %v\n", file)
		conf.Secrets = secrets
	}

	logging.ChangeLogLevel(conf.LogLevel)
	if strings.ToLower(conf.LogLevel) == "debug" {
		conf.PrintConfigToLog()
	}
}

func (conf *ApplicationConfig) PrintConfigToLog() {
	log.Infof("config-server: %v", conf)
}
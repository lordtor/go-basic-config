# GO Base Config module

Tis module used as base fo configuration apps.By default, it expands into the inside of the application. Also, module c reads a dictionary of secrets from the application directory by its `AppName` and extension `json`. 

## ENV parameters

* SPRING_CLOUD_CONFIG_URI - Config server URI
* APP_NAME - App name
* PROFILE_NAME - Profile name for app

## File parameters `application.yml`

``` GO
AppName       string `yaml:"app_name"`
ConfServerURI string `yaml:"conf_server_uri"`
LogLevel      string `yaml:"log_level"`
```

## Lint, Test & Coverage

1. Install linter

``` bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint
```

run check

``` bash
../../bin/golangci-lint run  --out-format checkstyle > checkstyle.xml
```

2. Run test

``` bash
go test -v -coverpkg=./... -coverprofile=profile.cov ./... -json > test_report.json
```

3. Coverage

``` bash
go tool cover -func profile.cov
```

## Extension module example

The base interface implements the following functions:

```GO
type Congig interface {
    GetSecretsFromJson() string
    GetParamsFromYml(path string)
    ParseCloudFile()
    ReloadConfig()
    PrintConfigToLog()
    UpdateSecrets()
}
```

Then, in the extension package, you can redefine some of the functions as follows:

```GO
package extend_config

import (
    "io/ioutil"
    "path/filepath"
    "strings"

    "github.com/imdario/mergo"
    "github.com/lordtor/go-basic-config"
    "github.com/lordtor/go-basic-couchdb_helper"
    "github.com/lordtor/go-logging"
    "gopkg.in/yaml.v2"
)

var (
    Log = logging.Log
)
// Create new struct for configuration & extend + couchdb
type C struct {
    base_config.ApplicationConfig `yaml:"app"` //base
    CouchDB                       couchdb_helper.CouchDB `yaml:"couchdb"` //extend
}

func (conf *C) GetParamsFromYml(path string) {
    if path == "" {
        path, _ = filepath.Abs("./application.yml")
    }
    Log.Info("[EXT:GetParamsFromYml]:: Load file: ", path)
    yamlFile, err := ioutil.ReadFile(path)
    if err != nil {
        Log.Error("[EXT:GetParamsFromYml]:: cannot open file: ", err)
    }
    err = yaml.Unmarshal(yamlFile, &conf)
    if err != nil {
        Log.Error("[EXT:GetParamsFromYml]:: cannot unmarshal data: ", err)
    }
    err = mergo.Merge(&C{}, conf)
    if err != nil {
        Log.Error("[EXT:GetParamsFromYml]:: cannot Merge data: ", err)
    }
}

func (conf *C) ReloadConfig() {
    Log.Info("[EXT:ReloadConfig]:: Start func ReloadConfig")
    conf.GetParamsFromYml("")
    if conf.ConfServerURI == "" {
        conf.ConfServerURI = base_config.GetValueByNameFromEnv("SPRING_CLOUD_CONFIG_URI")
    }
    Log.Info("[EXT:ReloadConfig]:: SPRING_CLOUD_CONFIG_URI", conf.ConfServerURI)
    if conf.ConfServerURI != "" {
        conf.ParseCloudFile()
    }
    secrets, file, err := conf.GetSecretsFromJson()
    if err != nil {
        Log.Error("[EXT:ReloadConfig]:: ", err)
    } else {
        Log.Infof("[EXT:ReloadConfig]:: Use credential's from different file %v\n", file)
        conf.Secrets = secrets
        conf.ReloadPassword()
    }
    logging.ChangeLogLevel(conf.LogLevel)
    if strings.ToLower(conf.LogLevel) == "debug" {
        conf.PrintConfigToLog()
    }
}

func (conf *C) ParseCloudFile() {
    Log.Info(conf.AppName, conf.ProfileName)
    rawBytes := base_config.FetchFileFromCloud(conf.AppName, conf.ProfileName, conf.ConfServerURI)
    err := yaml.Unmarshal(rawBytes, &conf)
    if err != nil {
        Log.Fatal("[EXT:parseCloudFile]:: ", err)
    }
    Log.Info("[EXT:parseCloudFile]:: parse cloud file: ", conf)
    err = mergo.Merge(&C{}, conf)
    if err != nil {
        Log.Fatal("[EXT:parseCloudFile]:: ", err)
    }
}
// New func for mount passwords in struct
func (conf *C) ReloadPassword() {
    for k, v := range conf.Secrets {
        if v != "" {
            switch k {
            case "couchdb_password":
                conf.CouchDB.Password = v
            }
        }
    }
}
```

Use in main.go

```GO
package main

import (
    "mon-api/extend_config"

    "gl.homecredit.ru/golib/logging"
    "gl.homecredit.ru/golib/version"
)

var (
    Conf            = extend_config.C{}
    Log             = logging.Log
    binVersion      = "0.1.5"
    aBuildNumber    = ""
    aBuildTimeStamp = ""
    aGitBranch      = ""
    aGitHash        = ""
)

func init() {
    version.InitVersion(binVersion, aBuildNumber, aBuildTimeStamp, aGitBranch, aGitHash)
    Log.Info(version.GetVersion())
    Conf.ReloadConfig()
}
...
```
package go_base_config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

var C_em = ApplicationConfig{}
var C_not_em = ApplicationConfig{
	AppName:     "base_config",
	ProfileName: "test",
	LogLevel:    "Debug",
}
var BinName = filepath.Base(os.Args[0])

func Path(a string) string {
	b, _ := filepath.Abs(a)
	return b
}
func TestExists(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"LICENSE", args{Path("LICENSE")}, true, false},
		{"LICENSE1", args{Path("LICENSE1")}, false, false},
		{"Error", args{"///"}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Exists(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}
func DeleteFile(f string) string {
	os.Remove(f)

	return f

}
func CreateFile(f string, data []byte) string {
	_ = ioutil.WriteFile(f, data, 0644)
	return f

}
func GetFileName(name, ext string) string {
	filename, _ := filepath.Abs(fmt.Sprintf("%s%s", name, ext))
	return filename
}
func TestApplicationConfig_GetSecretsFromJson(t *testing.T) {
	var data = Secrets{
		"password": "secret",
	}

	file, _ := json.MarshalIndent(data, "", " ")
	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		conf    *ApplicationConfig
		want    Secrets
		want1   string
		wantErr bool
	}{
		{"Read secret file by name exist", args{GetFileName(BinName, "")}, &C_em, data, CreateFile(GetFileName(BinName, ".json"), file), false},
		{"Read secret file by default name exist", args{""}, &C_em, data, CreateFile(GetFileName(BinName, ".json"), file), false},
		{"Read secret file by name not exist", args{"filename"}, &C_em, nil, GetFileName("filename", ".json"), false},
		{"Read secret file by name exist & not valid json", args{GetFileName("filename1", "")}, &C_em, nil, CreateFile(GetFileName("filename1", ".json"), []byte("file")), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.conf.GetSecretsFromJson(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplicationConfig.GetSecretsFromJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplicationConfig.GetSecretsFromJson() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ApplicationConfig.GetSecretsFromJson() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
	time.Sleep(3 * time.Second)
	DeleteFile(GetFileName(BinName, ".json"))
	DeleteFile(GetFileName("filename1", ".json"))

}

func TestApplicationConfig_GetParamsFromYml(t *testing.T) {
	type args struct {
		path string
	}
	y, _ := yaml.Marshal(C_not_em)
	CreateFile(GetFileName("params", ".yml"), y)
	CreateFile(GetFileName("params_not_valid", ".yml"), []byte("fwedfsfsd"))
	CreateFile(GetFileName("application", ".yml"), y)
	tests := []struct {
		name    string
		conf    *ApplicationConfig
		args    args
		wantErr bool
	}{
		{"Read params file by default name exist", &C_em, args{""}, false},
		{"Read params file by name exist", &C_em, args{GetFileName("params", ".yml")}, false},
		{"Read params file by name exist", &C_em, args{GetFileName("params_not_valid", ".yml")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.conf.GetParamsFromYml(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("ApplicationConfig.GetParamsFromYml() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	DeleteFile(GetFileName("params", ".yml"))
	DeleteFile(GetFileName("params_not_valid", ".yml"))
	DeleteFile(GetFileName("application", ".yml"))
}

func TestGetValueByNameFromEnv(t *testing.T) {
	type args struct {
		aName string
	}
	os.Setenv("Test", "Test")
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Exit", args{"Test"}, "Test"},
		{"NotExit", args{"Test1"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetValueByNameFromEnv(tt.args.aName); got != tt.want {
				t.Errorf("GetValueByNameFromEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchFileFromCloud(t *testing.T) {
	//want_got, _ := FetchFileFromCloud(C_not_em.AppName, C_not_em.ProfileName, C_not_em.ConfServerURI)
	//want_got_p, _ := FetchFileFromCloud(appName, "", confServerURI)
	type args struct {
		AppName       string
		ProfileName   string
		ConfServerURI string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		//{"Get sample config profile exist", args{AppName: C_not_em.AppName, ProfileName: C_not_em.ProfileName, ConfServerURI: C_not_em.ConfServerURI}, want_got, false},
		//{"Get sample config no ConfServerURI", args{AppName: C_not_em.AppName, ProfileName: C_not_em.ProfileName, ConfServerURI: ""}, nil, true},
		//{"Get sample config no profile", args{AppName: appName, ConfServerURI: confServerURI}, want_got_p, false},
		//{"Get sample config no profile & app name", args{ConfServerURI: C_not_em.ConfServerURI}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchFileFromCloud(tt.args.AppName, tt.args.ProfileName, tt.args.ConfServerURI)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFileFromCloud() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchFileFromCloud() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplicationConfig_ParseCloudFile(t *testing.T) {
	C_em.AppName = "base_config"
	C_em.ProfileName = ""
	C_em.LogLevel = "Debug"
	C_em.ConfServerURI = ""
	tests := []struct {
		name    string
		conf    *ApplicationConfig
		want    *ApplicationConfig
		wantErr bool
	}{
		{"empty url", &C_em, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.conf.ParseCloudFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplicationConfig.ParseCloudFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplicationConfig.ParseCloudFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplicationConfig_ReloadConfig(t *testing.T) {
	tests := []struct {
		name string
		conf *ApplicationConfig
	}{
		{"Base emty conf", &C_em},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.conf.ReloadConfig()
		})
	}
}

func TestApplicationConfig_PrintConfigToLog(t *testing.T) {
	tests := []struct {
		name string
		conf *ApplicationConfig
	}{
		{"Print config", &C_em},
		{"Print config2", &C_not_em},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.conf.PrintConfigToLog()
		})
	}
}

package configuration

import (
	"LiteCanary/internal/server"
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/viper"
)

var (
	defaultValues = map[string]any{
		"debug":            false,
		"databaselocation": ":memory:",
		"listener":         "127.0.0.1:8080",
		"basepath":         "/api/",
		"noregistration":   false,
		"publickey":        "",
		"privatekey":       "",
		"log":              "",
	}

	arguments = make(map[string]any)
)

func GetOptions() (*server.Options, error) {
	options := server.Options{
		NoRegistration:   defaultValues["noregistration"].(bool),
		Debug:            defaultValues["debug"].(bool),
		DatabaseLocation: defaultValues["databaselocation"].(string),
		Listener:         defaultValues["listener"].(string),
		BasePath:         defaultValues["basepath"].(string),
		PublicKey:        defaultValues["publickey"].(string),
		PrivateKey:       defaultValues["privatekey"].(string),
		Log:              defaultValues["log"].(string),
	}

	info, err := os.Stat("litecanary.conf")
	if err == nil && !info.IsDir() {
		viper.SetConfigFile("litecanary.conf")
		viper.SetConfigType("env")
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
		if err := viper.Unmarshal(&options); err != nil {
			return nil, err
		}
	}

	arguments["NoRegistrationP"] = flag.Bool("no-req", false, "disables registration")
	arguments["DebugP"] = flag.Bool("debug", false, "enables or disables debug information")
	arguments["DatabaseLocationP"] = flag.String("database", "", "database location (./test.db, :memory:)")
	arguments["ListenerP"] = flag.String("listener", "", "listener (127.0.0.1:8080)")
	arguments["BasePathP"] = flag.String("base", "", "base path for the api (/api/)")
	arguments["PublicKeyP"] = flag.String("cert", "", "public key for the rest api")
	arguments["PrivateKeyP"] = flag.String("key", "", "private key for the rest api")
	arguments["LogP"] = flag.String("log", "", "log file (disabled by default)")
	flag.Parse()
	for key, val := range arguments {
		valueType := fmt.Sprintf("%T", val)
		switch valueType {
		case "*bool":
			normal := *(val.(*bool))
			if normal {
				reflect.ValueOf(&options).Elem().FieldByName(key[0 : len(key)-1]).SetBool(normal)
			}
		case "*string":
			normal := *(val.(*string))
			if normal != "" {
				reflect.ValueOf(&options).Elem().FieldByName(key[0 : len(key)-1]).SetString(normal)
			}
		}
	}

	return &options, nil
}

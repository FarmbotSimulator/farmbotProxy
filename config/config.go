package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

var (
	cfgFile     = "/etc/farmbotproxy/farmbotproxy.yaml"
	cfgDir      = "/etc/farmbotproxy"
	cfgFileName = "farmbotproxy"
	cfgFileExt  = "yaml"
)

func Version() string {
	return version
}

func Config(overwrite bool) error {

	if _, err := os.Stat(cfgFile); err != nil || overwrite {
		if overwrite {
			log.Println("Overwriting config file")
		}
		if err := os.MkdirAll("/etc/farmbotproxy", os.ModePerm); err != nil {
			log.Fatal(err)
		}
		f, err := os.Create(cfgFile)

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		if _, err := f.WriteString(defaultConfig); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func RemoveIndex(s []string, index int) []string {
	ret := make([]string, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func GetConfig(keys ...string) (interface{}, error) {
	viper.SetConfigName(cfgFileName)
	viper.SetConfigType(cfgFileExt)
	viper.AddConfigPath(cfgDir)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	// fmt.Println(viper.Get("update"))
	// fmt.Println(viper.Get("PORT"))
	// fmt.Println(viper.Get("siteapi").(map[string]interface{})["database"].(map[string]interface{})["mariadb"])
	var (
		ret interface{}
		// depth = len(keys)
		index = 0
	)

	ret = viper.Get(keys[0])
	// fmt.Println(keys)
	keys = RemoveIndex(keys, 0)
	// fmt.Println(keys)
	for _, configKey := range keys {
		ret = ret.(map[string]interface{})[configKey]
		index++
	}
	// fmt.Println(ret)
	return ret, nil
}

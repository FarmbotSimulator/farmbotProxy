package farmbot

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/FarmbotSimulator/farmbotProxy/config"
	"github.com/vicanso/go-axios"
)

var paths = map[string]string{
	"device":       "device",
	"points":       "points",
	"savedGardens": "saved_gardens",
	"gardens":      "saved_gardens",
	"regimens":     "regimens",
	"sequences":    "sequences",
}

func Get(item string, token string) (interface{}, error) {
	pathItem := paths[item]
	if len(pathItem) == 0 {
		buf := &bytes.Buffer{}
		for _, rune := range item {
			if unicode.IsUpper(rune) {
				buf.WriteRune('_')
			}
			buf.WriteRune(rune)
		}
		pathItem = buf.String()
		pathItem = strings.ToLower(pathItem)
	}

	headers := make(http.Header)
	headers.Set("Accept", "application/json")
	headers.Set("Authorization", token)
	FARMBOTURLInterface, _ := config.GetConfig("FARMBOTURL")
	FARMBOTURL := FARMBOTURLInterface.(string)
	ins := axios.NewInstance(&axios.InstanceConfig{
		BaseURL:     FARMBOTURL + "/" + pathItem + "/",
		EnableTrace: true,
		Client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		Timeout: 10 * time.Second,
		Headers: headers,
		OnDone: func(config *axios.Config, resp *axios.Response, err error) {

		},
	})
	resp_, err := ins.Get("/")
	if err != nil {
		return false, fmt.Errorf(fmt.Sprintf("..."))
	}
	// buf, _ := json.Marshal(resp_.Config.HTTPTrace.Stats())
	// fmt.Println(string(buf))
	// status
	if resp_.Status != 200 {
		return false, fmt.Errorf("...")
	}
	return string(resp_.Data), nil
}

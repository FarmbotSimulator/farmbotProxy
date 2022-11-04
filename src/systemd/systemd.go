package systemd

import (
	"fmt"
	"log"
	"strings"

	"github.com/FarmbotSimulator/farmbotProxy/config"
	"github.com/kardianos/service"
)

type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	return nil
}
func (p *program) run() error {
	return nil
}
func (p *program) Stop(s service.Service) error {
	return nil
}

func getService(production bool) (service.Service, error) {
	env := "prod"
	if !production {
		env = "dev"
	}
	serviceNameInterface, err := config.GetConfig("service", env, "name")
	if err != nil {
		log.Fatal(err)
	}
	serviceName := serviceNameInterface.(string)
	serviceDisplayNameInterface, err := config.GetConfig("service", env, "displayname")
	if err != nil {
		log.Fatal(err)
	}
	serviceDisplayName := serviceDisplayNameInterface.(string)
	serviceDescriptionInterface, err := config.GetConfig("service", env, "displayname")
	if err != nil {
		log.Fatal(err)
	}
	serviceDescription := serviceDescriptionInterface.(string)

	arguments := make([]string, 0)
	arguments = append(arguments, "run")
	if production {

		arguments = append(arguments, "--prod")
	}

	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"
	svcConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceDisplayName,
		Description: serviceDescription,
		Dependencies: []string{
			"Requires=network.target",
			"After=network-online.target syslog.target"},
		Option:    options,
		Arguments: arguments,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	return s, err
}
func Start(production bool) {
	s, err := getService(production)
	if status, _ := s.Status(); status == 1 {
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	if err = s.Start(); err != nil {
		log.Fatal(err)
	}
}
func Stop(production bool) {
	s, err := getService(production)
	if err != nil {
		log.Fatal(err)
	}
	if err = s.Stop(); err != nil {
		log.Fatal(err)
	}
}
func Status() {

}
func Restart(production bool) {
	s, err := getService(production)
	if err != nil {
		log.Fatal(err)
	}
	if err = s.Restart(); err != nil {
		log.Fatal(err)
	}
}
func Install(production bool) {
	s, err := getService(production)
	if err != nil {
		log.Fatal(err)
	}
	if err = s.Install(); err != nil {
		if strings.Contains(fmt.Sprint(err), "already exists") {
			if err := s.Uninstall(); err != nil {
				log.Fatal(err)
			}
			s.Install()
		} else {
			log.Fatal(err)
		}
	}
}
func Uninstall(production bool) {
	s, err := getService(production)
	if err != nil {
		log.Fatal(err)
	}
	if err = s.Uninstall(); err != nil {
		log.Fatal(err)
	}
}

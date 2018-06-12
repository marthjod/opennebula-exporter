package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/marthjod/gocart/api"
	"github.com/marthjod/gocart/ocatypes"
	"github.com/marthjod/gocart/vmpool"
	yaml "gopkg.in/yaml.v2"
)

type exporterConfig struct {
	Namespace     string `yaml:"namespace"`
	ListenAddress string `yaml:"listen_address"`
}

type apiConfig struct {
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	Endpoint    string `yaml:"endpoint"`
	InsecureSSL bool   `yaml:"insecure_ssl"`
}

type vmNameRegexpLabel struct {
	Name   string `yaml:"name"`
	Regexp string `yaml:"regexp"`
}

type userTemplateLabel struct {
	Name          string `yaml:"name"`
	TemplateField string `yaml:"field"`
}

type config struct {
	Exporter           exporterConfig      `yaml:"exporter"`
	API                apiConfig           `yaml:"api"`
	VMNameRegexpLabels []vmNameRegexpLabel `yaml:"labels_vm_name"`
	UserTemplateLabels []userTemplateLabel `yaml:"labels_user_template"`
}

const metricsPath = "/metrics"

func addUserTemplateLabels(vm *ocatypes.Vm, labels []userTemplateLabel) string {
	var labelAttrs []string

	for _, label := range labels {
		field, err := vm.UserTemplate.Items.GetCustom(label.TemplateField)
		if err != nil {
			field = "unknown"
		}
		labelAttrs = append(labelAttrs, fmt.Sprintf(`%s=%q`, label.Name, field))
	}

	return buildString(labelAttrs)

}

// TODO compile label regexps only once
func addVMNameRegexpLabels(vm *ocatypes.Vm, labels []vmNameRegexpLabel) string {
	var labelAttrs []string

	for _, label := range labels {
		labelMatch, err := regexp.Compile(label.Regexp)
		if err != nil {
			labelAttrs = append(labelAttrs, fmt.Sprintf(`%s=%q`, label.Name, err))
			continue
		}

		if labelMatch.MatchString(vm.Name) {
			matches := labelMatch.FindStringSubmatch(vm.Name)
			// not checking for nil here since it matched before
			match := matches[len(matches)-1]
			labelAttrs = append(labelAttrs, fmt.Sprintf(`%s=%q`, label.Name, match))
		}
	}

	return buildString(labelAttrs)
}

func main() {
	var configFile = flag.String("config", "opennebula-exporter.yaml", "Config file")

	flag.Parse()

	c, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalln(err)
	}

	var cfg config
	err = yaml.Unmarshal(c, &cfg)
	if err != nil {
		log.Fatalln(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.API.InsecureSSL},
	}

	apiClient, err := api.NewClient(cfg.API.Endpoint, cfg.API.User, cfg.API.Password, tr)
	if err != nil {
		log.Fatalln(err)
	}

	var apiHost string
	apiURL, err := url.Parse(cfg.API.Endpoint)
	if err != nil {
		apiHost = cfg.API.Endpoint
	} else {
		apiHost = apiURL.Hostname()
	}

	http.HandleFunc(metricsPath, func(w http.ResponseWriter, r *http.Request) {

		// TODO query asynchronously if faster (!)
		vmPool := vmpool.NewVmPool()
		if err := apiClient.Call(vmPool); err != nil {
			log.Fatalln(err)
		}

		for _, vm := range vmPool.Vms {
			var b strings.Builder

			fmt.Fprintf(&b, `%s_vms{name=%q,id="%d",lcm_state=%q,api_host=%q`,
				cfg.Exporter.Namespace, vm.Name, vm.Id, vm.LCMState, apiHost)

			if len(cfg.VMNameRegexpLabels) > 0 {
				b.WriteString(addVMNameRegexpLabels(vm, cfg.VMNameRegexpLabels))
			}

			if len(cfg.UserTemplateLabels) > 0 {
				b.WriteString(addUserTemplateLabels(vm, cfg.UserTemplateLabels))
			}

			b.WriteString("} 1\n")
			w.Write([]byte(b.String()))
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>OpenNebula VM Exporter (WIP)</title></head>
			<body>
			<h1>OpenNebula VM Exporter (WIP)</h1>
			<p><a href="` + metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Listening on %s\n", cfg.Exporter.ListenAddress)
	err = http.ListenAndServe(cfg.Exporter.ListenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}

}

func buildString(a []string) string {
	if len(a) > 0 {
		var b strings.Builder
		b.WriteString(",")
		b.WriteString(strings.Join(a, ","))
		return b.String()
	}

	return ""
}

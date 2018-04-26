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

type label struct {
	Name  string `yaml:"name"`
	Match string `yaml:"match"`
}

type config struct {
	Exporter exporterConfig `yaml:"exporter"`
	API      apiConfig      `yaml:"api"`
	Labels   []label        `yaml:"labels"`
}

const metricsPath = "/metrics"

func addLabels(vm *ocatypes.Vm, labels []label) string {
	var labelAttrs []string

	for _, label := range labels {
		labelMatch, err := regexp.Compile(label.Match)
		if err != nil {
			labelAttrs = append(labelAttrs, fmt.Sprintf(`%s="%s"`, label.Name, err))
			continue
		}
		if labelMatch.MatchString(vm.Name) {
			labelAttrs = append(labelAttrs, fmt.Sprintf(`%s="%s"`, label.Name, labelMatch.FindString(vm.Name)))
		}
	}

	if len(labelAttrs) > 0 {
		var b strings.Builder
		b.WriteString(",")
		b.WriteString(strings.Join(labelAttrs, ","))
		return b.String()
	}

	return ""
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

		// TODO query asynchronously
		vmPool := vmpool.NewVmPool()
		if err := apiClient.Call(vmPool); err != nil {
			log.Fatalln(err)
		}

		for _, vm := range vmPool.Vms {
			var b strings.Builder

			fmt.Fprintf(&b, `%s_vms{name="%s",lcm_state="%s",api_host="%s"`,
				cfg.Exporter.Namespace, vm.Name, vm.LCMState, apiHost)

			if len(cfg.Labels) > 0 {
				b.WriteString(addLabels(vm, cfg.Labels))
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

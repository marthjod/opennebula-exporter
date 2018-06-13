package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/marthjod/gocart/api"
	"github.com/marthjod/gocart/vmpool"
	"github.com/marthjod/opennebula-exporter/config"
	"github.com/marthjod/opennebula-exporter/handler"
	"github.com/marthjod/opennebula-exporter/labeling"
	yaml "gopkg.in/yaml.v2"
)

const configFileEnvVar = "OPENNEBULA_EXPORTER_CONFIG"

func main() {
	configFile := os.Getenv(configFileEnvVar)
	if configFile == "" {
		log.Fatalf("env var %q not set", configFileEnvVar)
	}

	c, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	var cfg config.Config
	err = yaml.Unmarshal(c, &cfg)
	if err != nil {
		log.Fatalln(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.API.InsecureSSL},
	}

	apiClient, err := api.NewClient(cfg.API.Endpoint, cfg.API.User, cfg.API.Password, tr, 30*time.Second)
	if err != nil {
		log.Fatalln(err)
	}

	// TODO query asynchronously if faster (!)
	vmPool := vmpool.NewVMPool()
	if err := apiClient.Call(vmPool); err != nil {
		log.Fatalln(err)
	}

	if cfg.Exporter.WriteStdout {
		lines := labeling.AddLabels(cfg, vmPool)
		fmt.Print(lines)
		return
	}

	hdlr := handler.NewHandler(cfg, vmPool)
	http.Handle(cfg.Exporter.MetricsPath, hdlr)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>OpenNebula VM Exporter (WIP)</title></head>
			<body>
			<h1>OpenNebula VM Exporter (WIP)</h1>
			<p><a href="` + cfg.Exporter.MetricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Listening on %s\n", cfg.Exporter.ListenAddress)
	log.Fatal(http.ListenAndServe(cfg.Exporter.ListenAddress, nil))
}

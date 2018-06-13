package handler

import (
	"fmt"
	"net/http"

	"github.com/marthjod/gocart/vmpool"
	"github.com/marthjod/opennebula-exporter/labeling"

	"github.com/marthjod/opennebula-exporter/config"
)

type Handler interface {
	http.Handler
}

type handler struct {
	cfg    config.Config
	vmPool *vmpool.VMPool
}

func NewHandler(cfg config.Config, vmPool *vmpool.VMPool) Handler {
	return &handler{
		cfg:    cfg,
		vmPool: vmPool,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lines := labeling.AddLabels(h.cfg, h.vmPool)
	fmt.Fprintf(w, lines)
}

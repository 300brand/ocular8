package handler

import (
	"github.com/golang/glog"
	"os/exec"
)

type Handler struct {
	Args      map[string]string // map of [arg]configName
	Exec      string            // Execution string
	Frequency int               // How many times to run per hour
	Name      string            // Handler name
	NSQ       struct {
		Channel string // NSQ channel to listen on
		Topic   string // NSQ topic to listen on
	}

	dir string // Path to handler directory
}

var Args = make(map[string]string)

func (h *Handler) SetDir(dir string) {
	h.dir = dir
}

func (h Handler) Run(data []byte) (err error) {
	cmd := exec.Command(h.Exec)
	cmd.Dir = h.dir
	if err = cmd.Start(); err != nil {
		glog.Errorf("Could not start %s: %s", h.Exec, err)
		return
	}
	if err = cmd.Wait(); err != nil {
		glog.Errorf("Error encountered while running %s: %s", h.Exec, err)
		return
	}
	return
}

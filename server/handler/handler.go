package handler

import (
	"bytes"
	"github.com/300brand/ocular8/server/config"
	"github.com/golang/glog"
	"os"
	"os/exec"
	"text/template"
)

type Handler struct {
	Command   []string // Handler command
	Frequency int      // How many times to run per hour
	Name      string   // Handler name
	NSQ       struct {
		Channel string // NSQ channel to listen on
		Topic   string // NSQ topic to listen on
	}
	Options map[string]interface{} // Additional configs sent to central config

	dir string // Path to handler directory
}

type HandlerOptions struct {
	Config     *config.ConfigType
	Options    map[string]interface{}
	Data       string
	HandlerDir string
}

func (h *Handler) SetDir(dir string) {
	h.dir = dir
}

func (h *Handler) ParsedCmd(data string) (cmd []string) {
	cmd = make([]string, len(h.Command))
	buf := bytes.NewBuffer(make([]byte, 256))
	options := HandlerOptions{
		Config:     config.Config,
		Options:    h.Options,
		Data:       data,
		HandlerDir: h.dir,
	}

	for i, raw := range h.Command {
		t, err := template.New("").Parse(raw)
		if err != nil {
			glog.Warningf("template.Parse(%s): %s", raw, err)
			cmd[i] = raw
			continue
		}

		buf.Reset()
		if err = t.Execute(buf, options); err != nil {
			glog.Warningf("template.Execute(%s): %s", raw, err)
			cmd[i] = raw
			continue
		}

		cmd[i] = buf.String()
	}
	return
}

func (h Handler) Run(data string) (err error) {
	args := h.ParsedCmd(data)
	glog.Infof("%q", args)

	// if err = os.Chdir(h.dir); err != nil {
	// 	return
	// }

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err = cmd.Start(); err != nil {
		return
	}
	if err = cmd.Wait(); err != nil {
		return
	}
	return
}

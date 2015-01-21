package handler

import (
	"bytes"
	"github.com/300brand/ocular8/server/config"
	"github.com/golang/glog"
	"os/exec"
	"path/filepath"
	"text/template"
)

type Handler struct {
	Args      map[string][]string // map of [arg]configName
	Exec      string              // Execution string
	Frequency int                 // How many times to run per hour
	Name      string              // Handler name
	NSQ       struct {
		Channel string // NSQ channel to listen on
		Topic   string // NSQ topic to listen on
	}

	dir        string   // Path to handler directory
	parsedArgs []string // Pre-parsed arguments
}

func (h *Handler) SetDir(dir string) {
	h.dir = dir
}

func (h *Handler) ParsedArgs() (args []string) {
	if h.parsedArgs != nil {
		return h.parsedArgs
	}

	args = make([]string, 0, 16)
	buf := bytes.NewBuffer(make([]byte, 256))

	for k, vals := range h.Args {
		args = append(args, k)
		for _, raw := range vals {
			t, err := template.New("").Parse(raw)
			if err != nil {
				glog.Warningf("template.Parse(%s): %s", raw, err)
				args = append(args, raw)
				continue
			}

			if err = t.Execute(buf, config.Config); err != nil {
				glog.Warningf("template.Execute(%s): %s", raw, err)
				args = append(args, raw)
				continue
			}

			args = append(args, buf.String())
		}
	}

	h.parsedArgs = args
	return
}

func (h Handler) Run(data []byte) (err error) {
	absExec, err := filepath.Abs(filepath.Join(h.dir, h.Exec))
	if err != nil {
		return
	}

	cmd := exec.Command(absExec, h.ParsedArgs()...)
	cmd.Dir = h.dir
	if err = cmd.Start(); err != nil {
		return
	}
	if err = cmd.Wait(); err != nil {
		return
	}
	return
}

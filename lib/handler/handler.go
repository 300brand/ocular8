package handler

import (
	"bytes"
	"os"
	"os/exec"
	"text/template"

	"github.com/golang/glog"
)

type CommandOptions struct {
	Etcd string
	Data string
	Dir  string
}

type Handler struct {
	Command []string
	Dir     string
}

func New(dir string, command []string) (h *Handler) {
	return &Handler{
		Command: command,
		Dir:     dir,
	}
}

func (h Handler) ParsedCmd(etcd, data string) (cmd []string) {
	cmd = make([]string, len(h.Command))
	buf := bytes.NewBuffer(make([]byte, 256))
	options := CommandOptions{
		Etcd: etcd,
		Data: data,
		Dir:  h.Dir,
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

func (h Handler) Run(etcd, data string) (err error) {
	args := h.ParsedCmd(etcd, data)
	glog.Infof("%q", args)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = h.Dir
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

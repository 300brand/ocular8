package web

import (
	"github.com/golang/glog"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/300brand/ocular8/server/config"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	appDir := filepath.Join(config.Config.WebAssets, "app")
	appFiles := make([]string, 0, 16)
	filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".js" {
			rel, err := filepath.Rel(appDir, path)
			if err != nil {
				glog.Errorf("filepath.Rel(%s, %s): %s", appDir, path, err)
				return err
			}
			appFiles = append(appFiles, rel)
		}
		return nil
	})

	indexPath := filepath.Join(config.Config.WebAssets, "index.gohtml")
	t, err := template.ParseFiles(indexPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		glog.Errorf("template.ParseFiles(%s): %s", indexPath, err)
	}

	data := struct{ AppFiles []string }{appFiles}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		glog.Errorf("template.Execute: %s", err)
	}
}

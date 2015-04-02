package web

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/golang/glog"
)

func Close() {}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	assetsApp := filepath.Join(AssetsDir, "app")
	appDir, err := filepath.Abs(assetsApp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		glog.Errorf("filepath.Abs(%s): %s", assetsApp, err)
		return
	}
	appFiles := make([]string, 0, 16)
	filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			glog.Warningf("filepath.Walk(%s): %s", appDir, err)
			return nil
		}
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

	indexPath := filepath.Join(AssetsDir, "index.gohtml")
	t, err := template.ParseFiles(indexPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		glog.Errorf("template.ParseFiles(%s): %s", indexPath, err)
		return
	}

	data := struct{ AppFiles []string }{appFiles}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		glog.Errorf("template.Execute: %s", err)
		return
	}
}

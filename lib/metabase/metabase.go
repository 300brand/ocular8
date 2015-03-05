package metabase

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
)

var apiUrl *url.URL

func init() {
	apiUrl, _ = url.Parse(`http://metabase.moreover.com/api/v10/articles`)
}

// Fetches next set of articles from Metabase. If sequenceId is 0, performs
// initial request which returns the latest articles in Metabase
func Fetch(apikey, sequenceId, store string) (r *Response, err error) {
	query := make(url.Values)
	query.Set("key", apikey)
	query.Set("limit", "500")
	if sequenceId != "" {
		query.Set("sequence_id", sequenceId)
	}
	apiUrl.RawQuery = query.Encode()
	glog.Infof("Requesting %s", apiUrl)

	req, err := http.NewRequest("GET", apiUrl.String(), nil)
	if err != nil {
		return
	}
	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Content-Type", "text/xml;charset=UTF-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var body io.Reader
	if store != "" {
		f, err := os.Create(filepath.Join(store, time.Now().Format("20060102T150405.xml")))
		if err != nil {
			glog.Error(err)
			return nil, err
		}
		defer f.Close()
		body = io.TeeReader(resp.Body, f)
	} else {
		body = resp.Body
	}

	r = new(Response)
	if err = xml.NewDecoder(body).Decode(r); err != nil {
		return
	}

	if r.Status == "FAILURE" {
		err = fmt.Errorf("[%d] %s", r.MessageCode, r.DeveloperMessage)
	}
	return
}

package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Attr struct {
	Lni      string `xml:"head>script>lexisnexis>lni"`
	Dpsi     string `xml:"head>script>lexisnexis>dpsi"`
	Pub      string `xml:"head>script>lexisnexis>pub"`
	Pubtype  string `xml:"head>script>lexisnexis>pubtype"`
	Url      string `xml:"head>script>lexisnexis>url"`
	Language string `xml:"head>script>lexisnexis>language"`
	Section  string `xml:"head>script>lexisnexis>section"`
}

const (
	COLLECTION = "articles"
	TOPIC      = "article.id.extract.goose"
)

var (
	XMLPREFIX    = []byte("<?xml ")
	ITEMSUFFIX   = []byte("</NEWSITEM>")
	ENCODING_OLD = []byte("IBM1047")
	ENCODING_NEW = []byte("ISO-8859-1")
	DOCTYPE_NEW  = []byte(`<!DOCTYPE NEWSITEM SYSTEM "newsitem.dtd">`)
	DOCTYPES_OLD = [][]byte{
		[]byte(`<!DOCTYPE NEWSITEM PUBLIC "//LN//NEWSITEMv01-000//EN" "http://lnxhome/les/dsa/dtd/NEWSITEMv01-000.dtd">`),
		[]byte(`<!DOCTYPE NEWSITEM PUBLIC "//LN//NEWSITEMv01-000//EN" "http://lnxhome/lex/dsa/dtd/NEWSITEMv01-000.dtd">`),
	}
)

var (
	dsn          = flag.String("mongo", "mongodb://localhost:27017/ocular8", "Connection string to MongoDB")
	nsqdhttp     = flag.String("nsqdhttp", "http://localhost:4151", "NSQd HTTP address")
	toFile       = flag.String("o", "", "Save individual documents to this directory")
	xsltprocPath = flag.String("xsltproc", "/usr/bin/xsltproc", "Path to xsltproc binary")
	newsitemXslt = flag.String("xslt", "newsitem.xslt", "Path to newsitem.xslt")
	newsitemDTD  = flag.String("dtd", "newsitem.dtd", "Path to newsitem.dtd")
)

var (
	db     *mgo.Database
	nsqURL *url.URL
)

func articleExists(lni string) (a *types.Article) {
	result := db.C("articles").Find(bson.M{"lexisnexis.lni": lni})
	if n, err := result.Count(); err != nil || n == 0 {
		if err != nil {
			glog.Errorf("articleExists: %s", err)
		}
		return nil
	}
	a = new(types.Article)
	if err := result.One(a); err != nil {
		glog.Errorf("articleExists: %s", err)
		return nil
	}
	return
}

// Extracts each XML document from LexisNexis files
func chunks(filename string) (ch chan []byte, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}

	ch = make(chan []byte, 4)

	go func() {
		defer f.Close()

		r := bufio.NewReader(f)
		var buf bytes.Buffer
		for line, readErr := r.ReadBytes('\n'); readErr == nil; line, readErr = r.ReadBytes('\n') {
			buf.Write(bytes.TrimRight(line, "\r\n"))
			if !bytes.HasPrefix(buf.Bytes(), XMLPREFIX) {
				glog.Fatalf("Buffer does not have proper prefix (%s): %10.10q...", XMLPREFIX, buf.Bytes())
			}
			if !bytes.HasSuffix(buf.Bytes(), ITEMSUFFIX) {
				continue
			}
			chunk := make([]byte, buf.Len())
			copy(chunk, buf.Bytes())
			ch <- chunk
			buf.Reset()
		}
		close(ch)
	}()

	return
}

func createURL(attr *Attr) string {
	if attr.Url != "" {
		return attr.Url
	}
	return fmt.Sprintf("http://ocular8.com/article/%s", attr.Lni)
}

// Extracts attributes snugged inside transformed XML
func extractAttr(html []byte) (attr *Attr, err error) {
	attr = new(Attr)
	err = xml.Unmarshal(html, attr)
	return
}

func feedName(attr *Attr) (name string) {
	if attr.Section == "" {
		return "default"
	}
	name = strings.ToLower(attr.Section)
	if strings.HasPrefix(name, "pg") {
		// Some items use "Pg. \d+" for their section
		return "default"
	}
	if idx := strings.Index(name, ";"); idx > -1 {
		name[:idx]
	}
	name = strings.Replace(name, " ", "_", -1)
	name = regexp.MustCompile(`\W`).ReplaceAllString(name, "")
	return
}

func parentIds(attr *Attr) (pubid, feedid bson.ObjectId, err error) {
	pub := new(types.Pub)
	if err = db.C("pubs").Find(bson.M{"lexisnexis.dpsi": attr.Dpsi}).One(pub); err == mgo.ErrNotFound {
		pub.Id = bson.NewObjectId()
		pub.IsLexisNexis = true
		pub.Name = attr.Pub
		pub.LexisNexis = &types.LexisNexisPub{
			Name: attr.Pub,
			DSPI: attr.Dpsi,
			// Track left undefined
		}
	}
	return bson.NewObjectId(), bson.NewObjectId(), nil
}

// Performs all LexisNexis processing items in order
func process(filename string) (err error) {
	ch, err := chunks(filename)
	if err != nil {
		return
	}

	var attr *Attr
	for rawxml := range ch {
		var html []byte
		if html, err = transform(rawxml); err != nil {
			return
		}
		if attr, err = extractAttr(html); err != nil {
			return
		}
		glog.Infof("%s | %s", attr.Dpsi, attr.Section)
		continue
		if attr.Language != "en" {
			glog.Warningf("Ignoring article: Language is '%s'", attr.Language)
			continue
		}

		// If there's an existing article, update it's XML and HTML
		var ln *types.LexisNexisArticle
		article := articleExists(attr.Lni)
		if article != nil {
			article.LexisNexis.XML = rawxml
			article.HTML = html
			goto SaveArticle
		}
		// New article
		ln = &types.LexisNexisArticle{
			XML:      rawxml,
			Filename: filename,
			LNI:      attr.Lni,
			DPSI:     attr.Dpsi,
		}
		article = &types.Article{
			Id:           bson.NewObjectId(),
			Url:          createURL(attr),
			IsLexisNexis: true,
			HTML:         html,
			LexisNexis:   ln,
		}
		article.PubId, article.FeedId = parentIds(attr)
	SaveArticle:
		if err := save(article); err != nil {
			glog.Errorf("process->save(%s): %s", article.Id.Hex(), err)
		}
	}
	return
}

func save(a *types.Article) (err error) {
	return

	// doc := &Doc{
	// 	Id:       bson.NewObjectId(),
	// 	Filename: filepath.Base(filename),
	// 	XML:      data,
	// }
	// if err = db.C(COLLECTION).Insert(doc); err != nil {
	// 	glog.Errorf("db.C(%s).Insert({Id:%s, Filename:%s, XML:%d})", COLLECTION, doc.Id, doc.Filename, len(data))
	// 	return
	// }
	// if dir := *toFile; dir != "" {
	// 	var f *os.File
	// 	if f, err = os.Create(filepath.Join(dir, doc.Id.Hex()+".xml")); err != nil {
	// 		return
	// 	}
	// 	defer f.Close()
	// 	if _, err = f.Write(data); err != nil {
	// 		return
	// 	}
	// }
	// return doc.Id, nil
}

// Compiles ObjectIds into a single payload for NSQ mpub and sends to TOPIC
func sendIds(ids []bson.ObjectId) (err error) {
	// Generate payload for NSQd
	payload := make([]byte, 0, len(ids)*25)
	for _, id := range ids {
		payload = append(payload, []byte(id.Hex())...)
		payload = append(payload, '\n')
	}
	body := bytes.NewReader(payload)
	bodyType := "multipart/form-data"

	// Send payload to NSQd
	if _, err := http.Post(nsqURL.String(), bodyType, body); err != nil {
		glog.Fatalf("http.Post(%s): %s", nsqURL.String(), err)
	}
	return
}

// Transforms XML document to "HTML" document to pass along to goose later
func transform(rawxml []byte) (html []byte, err error) {
	absDTD, err := filepath.Abs(*newsitemDTD)
	if err != nil {
		return
	}

	rawxml = bytes.Replace(rawxml, ENCODING_OLD, ENCODING_NEW, 1)
	for _, d := range DOCTYPES_OLD {
		rawxml = bytes.Replace(rawxml, d, DOCTYPE_NEW, 1)
	}

	cmd := exec.Command(*xsltprocPath, "--nonet", "--nomkdir", "--nowrite", "--path", filepath.Dir(absDTD), *newsitemXslt, "-")
	stdout, stderr := bytes.NewBuffer(make([]byte, 0, 8*1024)), bytes.NewBuffer(make([]byte, 0, 8*1024))
	cmd.Stdin = bytes.NewReader(rawxml)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err = cmd.Run(); err != nil {
		return
	}

	if stderr.Len() > 0 {
		err = fmt.Errorf("There was an error while transforming. See logs")
		// glog.Error(stderr.String())
		return
	}

	html = stdout.Bytes()
	return
}

func main() {
	var err error
	flag.Parse()

	if flag.NArg() == 0 {
		glog.Error("No files given")
		os.Exit(1)
	}

	nsqURL, err = url.Parse(*nsqdhttp)
	if err != nil {
		glog.Fatalf("Error parsing %s: %s", *nsqdhttp, err)
	}
	nsqURL.Path = "/mpub"
	nsqURL.RawQuery = (url.Values{"topic": []string{TOPIC}}).Encode()

	s, err := mgo.Dial(*dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", *dsn, err)
	}
	defer s.Close()
	db = s.DB("")

	for _, filename := range flag.Args() {
		if err := process(filename); err != nil {
			glog.Errorf("process(%s): %s", filename, err)
		}
	}
}

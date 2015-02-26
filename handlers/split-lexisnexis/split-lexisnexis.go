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
	"sync"

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

	ch = make(chan []byte)

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
		name = name[:idx]
	}
	name = strings.Replace(name, " ", "_", -1)
	name = regexp.MustCompile(`\W`).ReplaceAllString(name, "")
	return
}

func parentIds(attr *Attr) (pubid, feedid bson.ObjectId, err error) {
	// Find or create pub
	pub := new(types.Pub)
	pQuery := bson.M{"lexisnexis.dpsi": attr.Dpsi}
	if err = db.C("pubs").Find(pQuery).One(pub); err == mgo.ErrNotFound {
		pub.Id = bson.NewObjectId()
		pub.Homepage = fmt.Sprintf("http://www.lexis-nexis.com/%s", attr.Dpsi)
		pub.IsLexisNexis = true
		pub.Name = attr.Pub
		pub.LexisNexis = &types.LexisNexisPub{
			Name: attr.Pub,
			DPSI: attr.Dpsi,
			// Track left undefined
		}
		if err = db.C("pubs").Insert(pub); err != nil {
			return
		}
		err = nil
	}
	if err != nil {
		return
	}
	// Feed time
	feed := new(types.Feed)
	section := feedName(attr)
	fQuery := bson.M{
		"pubid":              pub.Id,
		"lexisnexis.section": section,
	}
	if err = db.C("feeds").Find(fQuery).One(feed); err == mgo.ErrNotFound {
		feed.Id = bson.NewObjectId()
		feed.PubId = pub.Id
		feed.Url = fmt.Sprintf("http://www.lexis-nexis.com/%s/%s.xml", attr.Dpsi, section)
		feed.IsLexisNexis = true
		feed.LexisNexis = &types.LexisNexisFeed{
			Section: section,
			// Track left undefined
		}
		if err = db.C("feeds").Insert(feed); err != nil {
			return
		}
		err = nil
	}
	if err != nil {
		return
	}
	return pub.Id, feed.Id, nil
}

// Performs all LexisNexis processing items in order
func process(filename string) (err error) {
	ch, err := chunks(filename)
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	for rawxml := range ch {
		wg.Add(1)
		go func(rawxml []byte) {
			defer wg.Done()
			var html []byte
			var attr *Attr
			if html, err = transform(rawxml); err != nil {
				return
			}
			if attr, err = extractAttr(html); err != nil {
				return
			}
			if attr.Language != "en" {
				glog.Warningf("Ignoring article: Language is '%s'", attr.Language)
				return
			}
			glog.Infof("%s | %s", attr.Dpsi, feedName(attr))
			return

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
			article.PubId, article.FeedId, err = parentIds(attr)
			if err != nil {
				return
			}
		SaveArticle:
			if err := save(article); err != nil {
				glog.Errorf("process->save(%s): %s", article.Id.Hex(), err)
			}
		}(rawxml)
	}
	wg.Wait()
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

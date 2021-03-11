package main 

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"archive/zip"
	"encoding/xml"
	"strings"
	"log"
)

type OfficeCoreProperty struct {

	XLMName xml.Name `xml:"coreProperties"`
	Creator string `xml:"creator"`
	LastModifiedBy string `xml:"lastModifiedBy"`
}

type OfficeAppProperty struct {

	XLMName xml.Name `xml:"Properties"`
	Application string `xml:"Application"`
	Company string `xml:"Company"`
	Version string `xml:"AppVersion"`
}

var OfficeVersions = map[string]string {

	"16": "2016",
	"15": "2013",
	"14": "2010",
	"13": "2007",
	"11": "2003",
}

func GetMajorVersion(a *OfficeAppProperty) string {

	tokens := strings.Split(a.Version, ".")
	if len(tokens) < 2 {
		return "Unknown"
	}

	v, ok := OfficeVersions[tokens[0]]
	if !ok {
		return "Unknown"
	}
	return v
}

func process(f *zip.File, prop interface {}) error {

	rc, err := f.Open()
	if err != nil {
		return err 
	}
	defer rc.Close()

	if error := xml.NewDecoder(rc).Decode(&prop); error != nil {
		return error
	}
	return nil
}

func NewProperties(r *zip.Reader)(*OfficeCoreProperty, *OfficeAppProperty, error) {

	var coreProps OfficeCoreProperty
	var appProps OfficeAppProperty

	for _, f := range r.File {
		switch f.Name {
			case "docProps/core.xml":
				if err := process(f, &coreProps); err != nil {
					return nil, nil, err
				}
			case "docProps/app.xml":
				if err := process(f, &appProps); err != nil {
					return nil, nil, err
				}
			default:
				continue
		}
	}
	return &coreProps, &appProps, nil
}

func handler(i int, s *goquery.Selection) {
	url, ok := s.Find("a").Attr("href")
	if !ok {
		fmt.Println("Error trying to find the link") 
	}

	fmt.Printf("%d: %s\n", i, url)

	res,err := http.Get(url)
	if err != nil {
		fmt.Println("Error request")
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error within buf")
	}
	defer res.Body.Close()

	r, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		fmt.Println("Error zip reader")
	}

	cp, ap, err := NewProperties(r)
	if err != nil {
		fmt.Println("Error within new properties func")
	}

	log.Printf(
		"%25s %25s - %s %s\n",
		cp.Creator,
		cp.LastModifiedBy, 
		ap.Application, 
		GetMajorVersion(ap))
}

func main() {

	if len(os.Args) != 3 {
		log.Fatalln("Missing required argument.")
	}

	domain := os.Args[1]
	filetype := os.Args[2]

	q := fmt.Sprintf(
		"site:%s && filetype:%s && instreamset:(url title):%s",
		domain,
		filetype,
		filetype)

	search := fmt.Sprintf("https://www.bing.com/search?q=%s&qs=n&form=QBRE", url.QueryEscape(q))
	res, err := http.Get(search)
	if err != nil {
		log.Panicln(err)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Panicln(err)
	}
	defer res.Body.Close()

	s := "html body div#b_content ol#b_results li.b_algo h2"
	doc.Find(s).Each(handler)
}

package main

import (
	"code.google.com/p/go.net/html"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
)

// Commammand line flags definition
var link = flag.String("i", "", "Link to open")
var fileType = flag.String("e", "jpg,jpeg,gif,png", "Comma separated list of filetypes to fetch")
var debug = flag.Bool("d", false, "Enable debug")
var doGet = flag.Bool("g", false, "Download the links in current location")

var wg sync.WaitGroup

func GetAllLinks(data io.ReadCloser) (links []string, err error) {
	tokenizer := html.NewTokenizer(data)
	for {
		tokenizer.Next()
		token := tokenizer.Token()
		switch token.Type {
		case html.ErrorToken:
			return
		case html.EndTagToken:
		case html.CommentToken:
		case html.TextToken:
		case html.StartTagToken, html.SelfClosingTagToken:
			if *debug {
				log.Print("type ", token.Type)
				log.Print("data ", token.Data)
			}
			if token.Data == "a" {
				for _, a := range token.Attr {
					if a.Key == "href" {
						for _, ext := range strings.Split(*fileType, ",") {
							if strings.HasSuffix(a.Val, ext) {
								if strings.HasPrefix(a.Val, "//") {
									links = append(links, "http:"+a.Val)
								} else {
									links = append(links, a.Val)
								}
							}
						}
					}
				}
			}
		}
	}
	return
}

func DownloadToLocal(l string) {
	pieces := strings.Split(l, "/")
	resp, err := http.Get(l)
	if err != nil {
		log.Print(err)
		wg.Done()
		return
	}
	defer resp.Body.Close()
	out, err := os.Create(pieces[len(pieces)-1])
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	io.Copy(out, resp.Body)
	wg.Done()
}

func main() {
	// must be here for flag parsing to work!
	flag.Parse()
	// parallelize this thing up
	runtime.GOMAXPROCS(runtime.NumCPU())

	if len(*link) == 0 {
		log.Fatal("Please specify a link to fetch!")
	}

	if *debug {
		log.Print("link:", *link)
		log.Print("fileType:", *fileType)
		log.Print("debug:", *debug)
		log.Print("doGet:", *doGet)
		log.Print("remaining args:", flag.Args())
	}

	resp, err := http.Get(*link)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error fetching %s: %s", *link, err))
	}

	linkList, err := GetAllLinks(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if !*doGet {
		for _, l := range linkList {
			log.Print("Link: ", l)
		}
	} else {
		for _, l := range linkList {
			wg.Add(1)
			go DownloadToLocal(l)
		}
	}
	wg.Wait()
}

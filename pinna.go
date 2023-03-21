package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

/* func getsite (url string) {

} */

func main() {

	s := os.Args[1]
	// "https://home.treasury.gov"
	// "https://home.treasury.gov/news/press-releases/statements-remarks"
	u, err := url.Parse(s)
	if err != nil {
		fmt.Println(err)
	}

	domain := strings.Join([]string{u.Scheme, u.Host}, "://")
	subs := []string{}
	paths := strings.Split(u.Path, "/")
	subs = append(subs, paths[1:]...)
	walks := []string{}

	if len(subs) > 0 {
		for i := len(subs); i >= 1; i-- {
			walk := domain + "/" + strings.Join(subs[:i], "/")
			walks = append(walks, walk)
		}
	} else {
		fmt.Printf("No directory paths found in URL.\nOnly %s will be evaluated.\n", domain)
		walks = append(walks, domain)
	}

	// Manage TLS and redirects
	/* tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}
	} */

	client := &http.Client{
		// Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, v := range walks[0:] {
		fmt.Printf("\n%s\n\n", v)
		resp, err := client.Get(v)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer resp.Body.Close()

		// Output some header information
		if resp.StatusCode >= 300 && resp.StatusCode <= 399 {
			fmt.Printf("Status Code: %s (Redirects to: %s\n)", resp.Status, resp.Header.Get("Location"))
		} else {
			fmt.Printf("Status Code: %s\n", resp.Status)
		}
		headers := []string{"Content-Language", "Content-Type", "Last-Modified"}
		for _, i := range headers {
			h, head := resp.Header[i]
			if head {
				fmt.Printf("%s: %s\n", i, h[0])
			} else {
				fmt.Printf("%s: Not found in response\n", i)
			}
		}

		// Output some information about the body
		ct := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "text/html") {
			fmt.Printf("Content type is %s not html\n", ct)
		} else {
			tokzr := html.NewTokenizer(resp.Body)
			var hrefs int
			for {
				ttype := tokzr.Next()
				if ttype == html.ErrorToken {
					err := tokzr.Err()
					if err == io.EOF {
						break
					} else {
						fmt.Printf("Error: %s", err)
						continue
					}
				}
				// Get the page title and links
				if ttype == html.StartTagToken {
					token := tokzr.Token()
					if token.Data == "title" {
						ttype = tokzr.Next()
						if ttype == html.TextToken {
							fmt.Printf("Title: %s\n", tokzr.Token().Data)
						}
					} else {
						for _, i := range token.Attr {
							if i.Key == "href" {
								hrefs++
								//fmt.Println(i.Val)
							}
						}
					}
				}
			}
			fmt.Printf("hrefs: %d\n", hrefs)
		}
	}
}

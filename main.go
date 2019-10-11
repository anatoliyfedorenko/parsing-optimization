package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"golang.org/x/net/html"
)

func main() {

}

// CaseOne is when no go routines involved
func CaseOne(poolSize int) {
	for i := 1; i <= poolSize; i++ {
		_, err := getAds(
			fmt.Sprintf("https://krisha.kz/arenda/kvartiry/?page=%v", i))
		if err != nil {
			log.Fatal(err)
		}
	}
}

// CaseTwo is when go routines are involved but not reused
func CaseTwo(poolSize int) {
	var wg sync.WaitGroup

	for i := 1; i <= poolSize; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			_, err := getAds(
				fmt.Sprintf("https://krisha.kz/arenda/kvartiry/?page=%v", page))
			if err != nil {
				log.Fatal(err)
			}
		}(i)
	}

	wg.Wait()
}

// CaseThree is when go routines are involved and reused
func CaseThree(poolSize int) {
	var wg sync.WaitGroup
	pagesChan := make(chan int)
	go func() {
		for i := 1; i <= poolSize; i++ {
			pagesChan <- i
		}
		close(pagesChan)
	}()
	wg.Add(16)
	for i := 1; i <= 16; i++ {
		go func(pages chan int) {
			defer wg.Done()
			for {
				page, ok := <-pages
				if ok == false {
					break
				}
				_, err := getAds(fmt.Sprintf(
					"https://krisha.kz/arenda/kvartiry/?page=%v", page))
				if err != nil {
					log.Fatal(err)
				}
			}
		}(pagesChan)
	}
	wg.Wait()
}

func getAds(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	ads := []string{}

	// Recursively visit nodes in the parse tree
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "a-card__title " {
					ads = append(ads, n.FirstChild.Data)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	return ads, nil
}

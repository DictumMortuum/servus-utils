package scrape

import (
	"github.com/DictumMortuum/servus/pkg/w3m"
	"github.com/gocolly/colly/v2"
	"log"
	"net/http"
)

func ScrapeMeepleOnBoard() (map[string]any, []map[string]any, error) {
	store_id := int64(10)
	rs := []map[string]any{}

	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))

	collector := colly.NewCollector()
	collector.WithTransport(t)

	collector.OnHTML("div.product-small.purchasable", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".amount")

		var stock int

		if hasClass(e, "instock") {
			stock = 0
		} else if hasClass(e, "onbackorder") {
			stock = 1
		} else if hasClass(e, "out-of-stock") {
			stock = 2
		}

		item := map[string]any{
			"name":        e.ChildText(".name"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".attachment-woocommerce_thumbnail", "src"),
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.Request.AbsoluteURL(e.ChildAttr(".name a", "href")),
		}

		rs = append(rs, item)
	})

	collector.OnHTML("a.next", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Println("Visiting: " + link)
		local_link, _ := w3m.BypassCloudflare(link)
		collector.Visit(local_link)
	})

	local, err := w3m.BypassCloudflare("https://meepleonboard.gr/product-category/board-games")
	if err != nil {
		return nil, nil, err
	}

	collector.Visit(local)
	collector.Wait()

	return map[string]interface{}{
		"name":    "Meeple on Board",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

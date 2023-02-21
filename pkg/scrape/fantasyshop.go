package scrape

import (
	"github.com/DictumMortuum/servus/pkg/w3m"
	"github.com/gocolly/colly/v2"
	"log"
	"net/http"
)

func ScrapeFantasyShop() (map[string]any, []map[string]any, error) {
	store_id := int64(28)
	rs := []map[string]any{}

	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))

	collector := colly.NewCollector()
	collector.WithTransport(t)

	collector.OnHTML(".ty-column3", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".ty-price-num")

		item := map[string]any{
			"name":        e.ChildText(".product-title"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".ty-pict.cm-image", "src"),
			"stock":       0,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr(".ty-grid-list__image a", "href"),
		}

		rs = append(rs, item)
	})

	collector.OnHTML("a.ty-pagination__next", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Println("Visiting: " + link)
		local_link, _ := w3m.BypassCloudflare(link)
		collector.Visit(local_link)
	})

	local, err := w3m.BypassCloudflare("https://www.fantasy-shop.gr/epitrapezia-paihnidia-pazl/?features_hash=18-Y")
	if err != nil {
		return nil, nil, err
	}

	collector.Visit(local)
	collector.Wait()

	return map[string]interface{}{
		"name":    "Fantasy Shop",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

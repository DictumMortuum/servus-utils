package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeKaissaEu() (map[string]any, []map[string]any, error) {
	store_id := int64(6)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("www.kaissa.eu"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML("article.product", func(e *colly.HTMLElement) {
		var stock int
		raw_price := e.ChildText(".price")

		if childHasClass(e, ".add-to-cart input", "stock-update") {
			stock = 2
		} else {
			stock = 0
		}

		item := map[string]any{
			"name":        e.ChildText(".caption"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".photo a img", "src"),
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.Request.AbsoluteURL(e.ChildAttr(".photo a", "href")),
		}

		rs = append(rs, item)
	})

	collector.OnHTML(".next a", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://www.kaissa.eu/products/epitrapezia-kaissa")
	collector.Visit("https://www.kaissa.eu/products/epitrapezia-sta-agglika")
	collector.Wait()

	return map[string]interface{}{
		"name":    "Kaissa Eu",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

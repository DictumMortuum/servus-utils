package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeKaissaGames() (map[string]any, []map[string]any, error) {
	store_id := int64(9)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("kaissagames.com"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML("li.item.product-item", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".price")

		var stock int

		if e.ChildText(".release-date") != "" {
			stock = 1
		} else {
			if !childHasClass(e, "div.stock", "unavailable") {
				stock = 0
			} else {
				stock = 2
			}
		}

		item := map[string]any{
			"name":        e.ChildText(".name"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".product-image-photo", "src"),
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr(".name a", "href"),
		}

		rs = append(rs, item)
	})

	collector.OnHTML("a.next", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://kaissagames.com/b2c_gr/xenoglossa-epitrapezia.html")
	collector.Wait()

	return map[string]interface{}{
		"name":    "Kaissa Games",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

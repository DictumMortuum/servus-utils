package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeMeeplePlanet() (map[string]any, []map[string]any, error) {
	store_id := int64(7)
	rs := []map[string]any{}
	detected := 0

	collector := colly.NewCollector(
		colly.AllowedDomains("meeple-planet.com"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML("div.product-small.product-type-simple", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".amount")

		var stock int

		if hasClass(e, "out-of-stock") {
			stock = 2
		} else {
			if e.ChildText(".badge-inner") != "" {
				stock = 1
			} else {
				stock = 0
			}
		}

		item := map[string]any{
			"name":        e.ChildText(".name"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".attachment-woocommerce_thumbnail", "src"),
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr(".name a", "href"),
		}

		rs = append(rs, item)
		detected++
	})

	collector.OnHTML("a.next.page-number", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://meeple-planet.com/category/epitrapezia-paixnidia")
	collector.Visit("https://meeple-planet.com/category/pre-orders")
	collector.Wait()

	return map[string]interface{}{
		"name":    "Meeple Planet",
		"id":      store_id,
		"scraped": detected,
	}, rs, nil
}

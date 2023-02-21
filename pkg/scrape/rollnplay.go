package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeRollnplay() (map[string]any, []map[string]any, error) {
	store_id := int64(26)
	rs := []map[string]any{}
	detected := 0

	collector := colly.NewCollector(
		colly.AllowedDomains("rollnplay.gr"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML(".product.type-product", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".price ins .amount")

		if raw_price == "" {
			raw_price = e.ChildText(".price .amount")
		}

		var stock int

		if hasClass(e, "instock") {
			stock = 0
		} else if hasClass(e, "onbackorder") {
			stock = 1
		} else if hasClass(e, "outofstock") {
			stock = 2
		}

		item := map[string]any{
			"name":        e.ChildText(".heading-title"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".has-back-image img", "data-src"),
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr(".heading-title a", "href"),
		}

		rs = append(rs, item)
		detected++
	})

	collector.OnHTML(".woocommerce-pagination a.next", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://rollnplay.gr/?product_cat=all-categories")
	collector.Wait()

	return map[string]interface{}{
		"name":    "Roll n Play",
		"id":      store_id,
		"scraped": detected,
	}, rs, nil
}

package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeEfantasy() (map[string]any, []map[string]any, error) {
	store_id := int64(8)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("www.efantasy.gr"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML("div.product.product-box", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".product-price a strong")

		var stock int

		if e.Attr("data-label") == "" {
			stock = 2
		} else if e.Attr("data-label") == "preorder" {
			stock = 1
		} else {
			stock = 0
		}

		item := map[string]any{
			"name":        e.ChildText(".product-title"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".product-image a img", "src"),
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.Request.AbsoluteURL(e.ChildAttr(".product-title a", "href")),
		}

		rs = append(rs, item)
	})

	collector.OnHTML(".pagination a", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://www.efantasy.gr/en/products/%CE%B5%CF%80%CE%B9%CF%84%CF%81%CE%B1%CF%80%CE%AD%CE%B6%CE%B9%CE%B1-%CF%80%CE%B1%CE%B9%CF%87%CE%BD%CE%AF%CE%B4%CE%B9%CE%B1/sc-all")
	collector.Wait()

	return map[string]interface{}{
		"name":    "eFantasy",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

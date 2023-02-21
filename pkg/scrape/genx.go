package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeGenx() (map[string]any, []map[string]any, error) {
	store_id := int64(27)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("www.genx.gr"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML(".white_bg", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".txtSale")

		if raw_price == "" {
			raw_price = e.ChildText(".txtPrice")
		}

		raw_stock := e.ChildText(".txtOutOfStock")

		var stock int

		if raw_stock == "" {
			stock = 0
		} else {
			stock = 2
		}

		item := map[string]any{
			"name":        e.ChildText(".txtTitle"),
			"store_id":    store_id,
			"store_thumb": e.Request.AbsoluteURL(e.ChildAttr(".hover01 a img", "src")),
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.Request.AbsoluteURL(e.ChildAttr(".hover01 a", "href")),
		}

		rs = append(rs, item)
	})

	collector.OnHTML(".prevnext", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://www.genx.gr/index.php?page=0&act=viewCat&catId=60&prdsPage=45")
	collector.Wait()

	return map[string]interface{}{
		"name":    "Genx",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

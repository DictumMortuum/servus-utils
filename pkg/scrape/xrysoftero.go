package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeXrysoFtero() (map[string]any, []map[string]any, error) {
	store_id := int64(21)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("xrysoftero.gr"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML(".thumbnail-container", func(e *colly.HTMLElement) {
		url := e.ChildAttr(".cover-image", "src")
		if url == "" {
			return
		}

		raw_price := e.ChildText(".price")
		item := map[string]any{
			"name":        e.ChildText(".product-title"),
			"store_id":    store_id,
			"store_thumb": url,
			"stock":       0,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr("a.relative", "href"),
		}

		rs = append(rs, item)
	})

	collector.OnHTML("a.next", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://xrysoftero.gr/362-epitrapezia-paixnidia?resultsPerPage=48&q=%CE%9C%CE%AC%CF%81%CE%BA%CE%B1%5C-%CE%95%CE%BA%CE%B4%CF%8C%CF%84%CE%B7%CF%82-%CE%9A%CE%AC%CE%B9%CF%83%CF%83%CE%B1")
	collector.Wait()

	return map[string]interface{}{
		"name":    "Xryso Ftero",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

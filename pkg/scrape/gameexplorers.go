package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeGameExplorers() (map[string]any, []map[string]any, error) {
	store_id := int64(22)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("www.gameexplorers.gr"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML(".single-product-item", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".regular-price")
		item := map[string]any{
			"name":        e.ChildText("h2:nth-child(1)"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr("a:nth-child(1) > img:nth-child(1)", "src"),
			"stock":       0,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr("a:nth-child(1)", "href"),
		}

		rs = append(rs, item)
	})

	collector.OnHTML(".product-pagination > a", func(e *colly.HTMLElement) {
		if e.Attr("title") == "επόμενη σελίδα" {
			link := e.Attr("href")
			log.Println("Visiting: " + link)
			collector.Visit(link)
		}
	})

	collector.Visit("https://www.gameexplorers.gr/kartes-epitrapezia/epitrapezia-paixnidia/items-grid-date-desc-1-60.html")
	collector.Wait()

	return map[string]interface{}{
		"name":    "Game Explorers",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeGameRules() (map[string]any, []map[string]any, error) {
	store_id := int64(4)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("www.thegamerules.com"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML(".product-layout", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".price-new")

		if raw_price == "" {
			raw_price = e.ChildText(".price-normal")
		}

		var stock int

		switch e.ChildText(".c--stock-label") {
		case "Εκτός αποθέματος":
			stock = 2
		case "Άμεσα Διαθέσιμο":
			stock = 0
		default:
			stock = 1
		}

		item := map[string]any{
			"name":        e.ChildText(".name"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".product-img div img", "data-src"),
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr(".name a", "href"),
		}

		log.Println(item["store_thumb"])

		rs = append(rs, item)
	})

	collector.OnHTML("a.next", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://www.thegamerules.com/epitrapezia-paixnidia?fa132=Board%20Game%20Expansions")
	collector.Visit("https://www.thegamerules.com/epitrapezia-paixnidia?fa132=Board%20Games")
	collector.Wait()

	return map[string]interface{}{
		"name":    "Game Rules",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

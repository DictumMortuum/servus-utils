package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeVgames() (map[string]any, []map[string]any, error) {
	store_id := int64(5)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("store.v-games.gr"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML("li.product.type-product", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".price")

		var stock int

		if hasClass(e, "instock") {
			stock = 0
		} else if hasClass(e, "onbackorder") {
			stock = 1
		} else if hasClass(e, "outofstock") {
			stock = 2
		}

		item := map[string]any{
			"name":        e.ChildText(".woocommerce-loop-product__title"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".woocommerce-loop-product__link img", "src"),
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr(".woocommerce-LoopProduct-link", "href"),
		}

		rs = append(rs, item)
	})

	collector.OnHTML(".woocommerce-pagination a.next", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://store.v-games.gr/category/board-games")
	collector.Wait()

	return map[string]interface{}{
		"name":    "Vgames",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

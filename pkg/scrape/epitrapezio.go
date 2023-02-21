package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeEpitrapezio() (map[string]any, []map[string]any, error) {
	store_id := int64(15)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("epitrapez.io"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML("li.product.type-product", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".price ins .amount")

		if raw_price == "" {
			raw_price = e.ChildText(".price .amount")
		}

		var stock int

		if e.ChildText("a.add_to_cart_button") != "" {
			stock = 0
		} else {
			stock = 2
		}

		item := map[string]any{
			"name":        e.ChildText(".woocommerce-loop-product__title"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".epz-product-thumbnail img", "data-src"),
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr(".woocommerce-LoopProduct-link", "href"),
		}

		rs = append(rs, item)
	})

	collector.OnHTML(".woocommerce-pagination a.next", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://epitrapez.io/product-category/epitrapezia/?Stock=allstock")
	collector.Wait()

	return map[string]interface{}{
		"name":    "epitrapezio",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

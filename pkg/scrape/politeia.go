package scrape

import (
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapePoliteia() (map[string]any, []map[string]any, error) {
	store_id := int64(12)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("www.politeianet.gr"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML(".browse-page-block", func(e *colly.HTMLElement) {
		raw_price := e.ChildText(".productPrice")
		if raw_price == "" {
			return
		}

		item := map[string]any{
			"name":        e.ChildText(".browse-product-title"),
			"store_id":    store_id,
			"store_thumb": e.ChildAttr(".browseProductImage", "src"),
			"stock":       0,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr(".browse-product-title", "href"),
		}

		rs = append(rs, item)
	})

	collector.OnHTML("a.pagenav", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		log.Println("Visiting: " + link)
		collector.Visit(link)
	})

	collector.Visit("https://www.politeianet.gr/index.php?option=com_virtuemart&category_id=948&page=shop.browse&subCatFilter=-1&langFilter=-1&pubdateFilter=-1&availabilityFilter=-1&discountFilter=-1&priceFilter=-1&pageFilter=-1&Itemid=721&limit=20&limitstart=0")
	collector.Wait()

	return map[string]interface{}{
		"name":    "Politeia",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

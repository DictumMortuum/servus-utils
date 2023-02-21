package scrape

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	// "strconv"
	"strings"
	"unicode"
)

func ScrapeMysteryBay() (map[string]any, []map[string]any, error) {
	store_id := int64(3)
	rs := []map[string]any{}

	collector := colly.NewCollector(
		colly.AllowedDomains("www.mystery-bay.com"),
		user_agent,
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML("li[data-hook=product-list-grid-item]", func(e *colly.HTMLElement) {
		raw_price := e.ChildText("span[data-hook=product-item-price-to-pay]")
		raw_url := e.ChildAttr("img", "src")
		urls := getURL(raw_url)

		url := ""
		if len(urls) > 0 {
			candidate := urls[0]
			filtered := strings.Split(candidate, "/v1/fill")
			url = filtered[0]
		}
		// style="background-image:url(https://static.wixstatic.com/media/9dcd7c_df5e66ff7168447ab10021bfa739a4cc~mv2.png/v1/fill/w_100,h_100,al_c,usm_0.66_1.00_0.01/9dcd7c_df5e66ff7168447ab10021bfa739a4cc~mv2.png);background-size:contain" data-hook="">
		var stock int

		if e.ChildText("span[data-hook=product-item-ribbon]") == "PRE-ORDER" {
			stock = 1
		} else {
			if e.ChildAttr("button[data-hook=product-item-add-to-cart-button]", "aria-disabled") == "true" {
				stock = 2
			} else {
				stock = 0
			}
		}

		item := map[string]any{
			"name":        e.ChildText("h3"),
			"store_id":    store_id,
			"store_thumb": url,
			"stock":       stock,
			"price":       getPrice(raw_price),
			"url":         e.ChildAttr("a[data-hook=product-item-container]", "href"),
		}

		rs = append(rs, item)
	})

	// collector.OnHTML("a.skOBQqy", func(e *colly.HTMLElement) {
	// 	page := strings.Split(e.Attr("data-hook"), "-")

	// 	if len(page) > 1 {
	// 		l, _ := strconv.Atoi(page[1])

	// 		for i := 1; i <= l; i++ {
	// 			link := fmt.Sprintf("%s%d", getPage(e.Request.AbsoluteURL("")), i)
	// 			log.Println("Visiting: " + link)
	// 			collector.Visit(link)
	// 		}
	// 	}
	// })

	for i := 1; i < 20; i++ {
		log.Println("Visiting: ", i)
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/diaxeirisis-poron?page=%d", i))
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/stratigikis?page=%d", i))
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/fantasias?page=%d", i))
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/mystirioy-tromoy?page=%d", i))
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/paixnidia-me-miniatoyres-dungeon-cr?page=%d", i))
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/oikogeneiaka?page=%d", i))
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/tis-pareas?page=%d", i))
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/paixnidia-me-kartes-zaria?page=%d", i))
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/lcg?page=%d", i))
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/war-games?page=%d", i))
		collector.Visit(fmt.Sprintf("https://www.mystery-bay.com/pre-orders?page=%d", i))
	}

	collector.Wait()

	return map[string]interface{}{
		"name":    "Mystery Bay",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

func getPage(url string) string {
	return strings.TrimRightFunc(url, func(r rune) bool {
		return unicode.IsNumber(r)
	})
}

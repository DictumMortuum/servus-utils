package scrape

import (
	"github.com/DictumMortuum/servus/pkg/w3m"
	"github.com/gocolly/colly/v2"
	"log"
	"net/http"
	"strings"
)

func ScrapeCrystalLotus() (map[string]any, []map[string]any, error) {
	store_id := int64(24)
	rs := []map[string]any{}

	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))

	collector := colly.NewCollector(
		user_agent,
		colly.CacheDir("/tmp"),
	)
	collector.WithTransport(t)

	collector.OnHTML(".grid__item", func(e *colly.HTMLElement) {
		link := e.ChildAttr(".motion-reduce", "src")
		if strings.HasPrefix(link, "//") {
			link = "https:" + link
		}

		raw_price := e.ChildText(".price__sale")
		item := map[string]any{
			"name":        e.ChildText(".card-information__text"),
			"store_id":    store_id,
			"store_thumb": link,
			"stock":       0,
			"price":       getPrice(raw_price),
			"url":         "https://crystallotus.eu" + e.ChildAttr("a.card-information__text", "href"),
		}

		rs = append(rs, item)
	})

	collector.OnHTML(".pagination__list li:last-child a", func(e *colly.HTMLElement) {
		link := "https://crystallotus.eu" + e.Attr("href")
		log.Println("Visiting: " + link)
		local_link, _ := w3m.BypassCloudflare(link)
		collector.Visit(local_link)
	})

	local, err := w3m.BypassCloudflare("https://crystallotus.eu/collections/board-games")
	if err != nil {
		return nil, nil, err
	}

	collector.Visit(local)
	collector.Wait()

	return map[string]interface{}{
		"name":    "Crystal Lotus",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

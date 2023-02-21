package scrape

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

type madnessRoot struct {
	XMLName  xml.Name         `xml:"products"`
	Products []madnessProduct `xml:"product"`
}

type madnessProduct struct {
	XMLName  xml.Name `xml:"product"`
	SKU      string   `xml:"id"`
	Name     string   `xml:"title"`
	ThumbUrl string   `xml:"image_link"`
	Price    string   `xml:"price"`
	Stock    string   `xml:"availability"`
	Link     string   `xml:"link"`
}

func madnessAvailbilityToStock(s string) int {
	switch s {
	case "in stock":
		return 0
	case "on backorder":
		return 1
	case "out of stock":
		return 2
	default:
		return 2
	}
}

func ScrapeBoardsOfMadness() (map[string]any, []map[string]any, error) {
	store_id := int64(16)
	rs := []map[string]any{}
	detected := 0

	link := "https://boardsofmadness.com/wp-content/uploads/woo-product-feed-pro/xml/sVVFMsJLyEEtvbil4fbIOdm8b4ha7ewz.xml"
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, nil, err
	}

	conn := &http.Client{}
	resp, err := conn.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	root := madnessRoot{}
	err = xml.Unmarshal(body, &root)
	if err != nil {
		return nil, nil, err
	}

	for _, item := range root.Products {
		item := map[string]any{
			"name":        item.Name,
			"store_id":    store_id,
			"store_thumb": item.ThumbUrl,
			"stock":       madnessAvailbilityToStock(item.Stock),
			"price":       getPrice(item.Price),
			"url":         item.Link,
		}
		rs = append(rs, item)
		detected++
	}

	return map[string]interface{}{
		"name":    "Boards of Madness",
		"id":      store_id,
		"scraped": detected,
	}, rs, nil
}

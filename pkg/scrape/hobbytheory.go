package scrape

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

type hobbyRoot struct {
	XMLName xml.Name      `xml:"mywebstore"`
	Store   hobbyProducts `xml:"products"`
}

type hobbyProducts struct {
	XMLName  xml.Name  `xml:"products"`
	Products []product `xml:"product"`
}

type product struct {
	XMLName      xml.Name `xml:"product"`
	SKU          string   `xml:"id"`
	Name         string   `xml:"name"`
	ThumbUrl     string   `xml:"image"`
	Category     string   `xml:"category"`
	Price        string   `xml:"price_with_vat"`
	Stock        string   `xml:"instock"`
	Availability string   `xml:"availability"`
	Link         string   `xml:"link"`
}

func ScrapeHobbyTheory() (map[string]any, []map[string]any, error) {
	store_id := int64(23)
	rs := []map[string]any{}

	link := "https://feed.syntogether.com/skroutz/xml?shop=hobbytheory.myshopify.com"
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

	payload := hobbyRoot{}
	err = xml.Unmarshal(body, &payload)
	if err != nil {
		return nil, nil, err
	}

	categories := []string{
		"Επιτραπέζια Παιχνίδια Οικογενειακά",
		"Επιτραπέζια Παιχνίδια Παρέας",
		"Επιτραπέζια Παιχνίδια Πολέμου",
		"Επιτραπέζια Παιχνίδια Στρατηγικής",
		"Θεματικά Επιτραπέζια Παιχνίδια",
	}

	for _, item := range payload.Store.Products {
		for _, category := range categories {
			if item.Category == category {
				var stock int

				if item.Stock == "Y" {
					stock = 0
				} else {
					stock = 2
				}

				item := map[string]any{
					"name":        item.Name,
					"store_id":    store_id,
					"store_thumb": item.ThumbUrl,
					"stock":       stock,
					"price":       getPrice(item.Price),
					"url":         item.Link,
				}

				rs = append(rs, item)
			}
		}
	}

	return map[string]interface{}{
		"name":    "Hobby Theory",
		"id":      store_id,
		"scraped": len(rs),
	}, rs, nil
}

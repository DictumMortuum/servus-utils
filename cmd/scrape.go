package main

import (
	"context"
	"fmt"
	rofi "github.com/DictumMortuum/gofi"
	"github.com/DictumMortuum/servus-utils/pkg/scrape"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var (
	Cfg Config
)

func scrapeSingle(db *sqlx.DB, f func() (map[string]any, []map[string]any, error), ignored []string) error {
	metadata, rs, err := f()
	if err != nil {
		return err
	}

	for _, item := range rs {
		if val, ok := item["name"]; ok {
			ignore := scrape.Ignore(ignored, val.(string))
			if !ignore {
				id, err := scrape.PriceExists(db, item)
				if err != nil {
					return err
				}

				if id != nil {
					log.Println(item["name"], "is mapped to id ", id)
					err = scrape.UpdatePrice(db, id, item)
					if err != nil {
						return err
					}
				} else {
					log.Println("inserting cached price", item["name"])
					err := scrape.InsertCachedPrice(db, item)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	log.Println(metadata)
	return nil
}

func main() {
	loader := confita.NewLoader(
		file.NewBackend("/etc/conf.d/servusrc.yml"),
	)

	err := loader.Load(context.Background(), &Cfg)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Connect("mysql", Cfg.Databases["mariadb"])
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db_ignored, err := scrape.GetIgnored(db)
	if err != nil {
		log.Fatal(err)
	}
	ignored := scrape.SetupIgnored(db_ignored)

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name: "scrape",
				Action: func(ctx *cli.Context) error {
					opts := rofi.GofiOptions{
						Description: "scraper",
					}

					scrapers, err := rofi.FromInterface(&opts, scrape.Scrapers)
					if err != nil {
						return err
					}

					for _, val := range scrapers {
						if f, ok := scrape.Scrapers[val].(func() (map[string]any, []map[string]any, error)); ok {
							err := scrapeSingle(db, f, ignored)
							if err != nil {
								return err
							}
						}
					}

					return nil
				},
			},
			{
				Name: "ignore",
				Action: func(ctx *cli.Context) error {
					prices, err := scrape.GetPrices(db)
					if err != nil {
						return err
					}

					for _, price := range prices {
						if scrape.Ignore(ignored, price.Name) {
							err = scrape.IgnorePrice(db, price.Id)
							if err != nil {
								return err
							}
							// fmt.Println(price.Name)
						}
					}

					cached_prices, err := scrape.GetCachedPrices(db)
					if err != nil {
						return err
					}

					deleted := 0
					for _, price := range cached_prices {
						if scrape.Ignore(ignored, price.Name) {
							err = scrape.DeleteCachedPrice(db, price.Id)
							if err != nil {
								return err
							}
							deleted++
							// fmt.Println(price.Name)
						}
					}

					fmt.Println("deleted:", deleted)

					return nil
				},
			},
			{
				Name: "clean",
				Action: func(ctx *cli.Context) error {
					opts := rofi.GofiOptions{
						Description: "scraper",
					}

					scrapers, err := rofi.FromInterface(&opts, scrape.IDs)
					if err != nil {
						return err
					}

					for _, val := range scrapers {
						if id, ok := scrape.IDs[val].(int); ok {
							err := scrape.CleanCachedPrices(db, id)
							if err != nil {
								return err
							}
							log.Println("Cleaning", val, "...")
						}
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

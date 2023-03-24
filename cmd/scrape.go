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

func scrapeSingle(db *sqlx.DB, f func() (map[string]any, []map[string]any, error), ign []scrape.Ignored) error {
	metadata, rs, err := f()
	if err != nil {
		return err
	}
	cached := 0
	updated := 0
	unchanged := 0
	invalid := 0
	ignored := 0

	for _, item := range rs {
		if val, ok := item["name"]; ok {
			ignore := scrape.Ignore(ign, val.(string), metadata["id"].(int64))
			if !ignore {
				id, boardgame_id, err := scrape.PriceExists(db, item)
				if err != nil {
					return err
				}

				item["id"] = id
				item["boardgame_id"] = boardgame_id

				if id != nil {
					err = scrape.UpdatePrice(db, item)
					if err != nil {
						return err
					}

					mapping_ok, err := scrape.InsertMapping(db, item)
					if err != nil {
						return err
					}

					history_ok, err := scrape.InsertHistories(db, item)
					if err != nil {
						return err
					}

					if mapping_ok || history_ok {
						log.Print(item["name"], "is mapped to id ", id, "...", boardgame_id)
						updated++
					} else {
						unchanged++
					}
				} else {
					if item["name"] != "" {
						cached_ok, err := scrape.InsertCachedPrice(db, item)
						if err != nil {
							return err
						}

						if cached_ok {
							log.Println("inserting cached price", item["name"])
							cached++
						}
					} else {
						invalid++
					}
				}
			} else {
				// log.Println("ignoring", item["name"])
				ignored++
			}
		}
	}

	metadata["cached"] = cached
	metadata["updated"] = updated
	metadata["unchanged"] = unchanged
	metadata["invalid"] = invalid
	metadata["ignored"] = ignored
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

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name: "scrape",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "store",
						Value: "",
					},
				},
				Action: func(ctx *cli.Context) error {
					var scrapers []string
					scraper := ctx.String("store")
					if scraper != "" {
						scrapers = []string{scraper}
					} else {
						opts := rofi.GofiOptions{
							Description: "scraper",
						}

						scrapers, err = rofi.FromInterface(&opts, scrape.Scrapers)
						if err != nil {
							return err
						}
					}

					for _, val := range scrapers {
						if f, ok := scrape.Scrapers[val].(func() (map[string]any, []map[string]any, error)); ok {
							err := scrapeSingle(db, f, db_ignored)
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
						if scrape.Ignore(db_ignored, price.Name, price.StoreId) {
							err = scrape.IgnorePrice(db, price.Id)
							if err != nil {
								return err
							}
						}
					}

					cached_prices, err := scrape.GetCachedPrices(db)
					if err != nil {
						return err
					}

					deleted := 0
					for _, price := range cached_prices {
						if scrape.Ignore(db_ignored, price.Name, price.StoreId) {
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
			{
				Name: "delete",
				Action: func(ctx *cli.Context) error {
					err := scrape.DeleteCachedPrices(db)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name: "insert",
				Action: func(ctx *cli.Context) error {
					_, err := scrape.CreatePricesFromCachedPrices(db)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name: "wrap",
				Action: func(ctx *cli.Context) error {
					rs, err := scrape.GetNewlyMappedPrices(db)
					if err != nil {
						return err
					}

					for _, item := range rs {
						err = scrape.MapPrice(db, item)
						if err != nil {
							return err
						}

						mapping_ok, err := scrape.InsertMapping(db, item)
						if err != nil {
							return err
						}

						history_ok, err := scrape.InsertHistories(db, item)
						if err != nil {
							return err
						}

						if mapping_ok || history_ok {
							log.Print(item.Name, "is mapped to id ", item.Id, "...", item.BoardgameId)
						} else {
							log.Println(item)
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

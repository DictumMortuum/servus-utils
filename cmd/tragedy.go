package main

import (
	"context"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
	"unicode"
)

var (
	Cfg Config
)

func scrape(c *cli.Context) error {
	author_id := c.Int("author_id")

	collector := colly.NewCollector(
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML(".library .row", func(e *colly.HTMLElement) {
		title := e.ChildText(".span4 .well h2")
		//fmt.Println(e.ChildAttr(".span4 .well h2 a", "href"))

		e.ForEach(".span8 .well .info li", func(idx int, tragedy *colly.HTMLElement) {
			tragedy.ForEach(".unstyled li", func(idx2 int, section *colly.HTMLElement) {
				_section := tragedy.ChildText("h4")
				part := section.ChildText("a")
				url := section.ChildAttr("a", "href")

				// fmt.Println(, section.ChildText("a"), section.ChildAttr("a", "href"))
				err := saveTragedy(title, _section, part, url)
				if err != nil {
					log.Fatal(err)
				}
			})
		})
	})

	collector.Visit(fmt.Sprintf("https://www.greek-language.gr/digitalResources/ancient_greek/library/index.html?start=0&author_id=%d", author_id))
	collector.Visit(fmt.Sprintf("https://www.greek-language.gr/digitalResources/ancient_greek/library/index.html?start=5&author_id=%d", author_id))
	collector.Visit(fmt.Sprintf("https://www.greek-language.gr/digitalResources/ancient_greek/library/index.html?start=10&author_id=%d", author_id))
	collector.Wait()
	return nil
}

func removeDigits(r rune) rune {
	if unicode.IsDigit(r) {
		r = -1
	}
	return r
}

func analyze(c *cli.Context) error {
	text_id := c.Int("text_id")
	page := c.Int("page")

	collector := colly.NewCollector(
		colly.CacheDir("/tmp"),
	)

	collector.OnHTML("#part p", func(e *colly.HTMLElement) {
		names := Unique(e.ChildTexts("span.name"))
		fmt.Println(names)
		current_name := ""

		for i, raw := range strings.Split(e.Text, "\n") {
			line := strings.Map(removeDigits, raw)

			if len(line) > 0 {
				flag := false
				for _, name := range names {
					if strings.HasPrefix(line, name) {
						flag = true
						current_name = name
					}
				}

				if flag {
					fmt.Println(i, flag, line)
				} else {
					fmt.Println(i, flag, current_name, line)
				}

			}
		}
		// fmt.Println(e.Text)
	})

	collector.Visit(fmt.Sprintf("https://www.greek-language.gr/digitalResources/ancient_greek/library/browse.html?text_id=%d&page=%d", text_id, page))
	collector.Wait()
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

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name: "scrape",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "author_id",
						Value: 102,
					},
				},
				Action: scrape,
			},
			{
				Name: "analyze",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "text_id",
						Value: 114,
					},
					&cli.IntFlag{
						Name:  "page",
						Value: 1,
					},
				},
				Action: analyze,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// Package main ...
package main

import (
	"context"
	"github.com/DictumMortuum/servus/pkg/models"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/ysmood/gson"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var (
	Cfg Config
)

func parseVoice(page *rod.Page, stats *models.Modem) {
	status := page.MustElement(`div.table-row:nth-child(2) > div:nth-child(3)`).MustText()
	stats.VoipStatus = status == "Up"
}

func parseDSL(page *rod.Page, stats *models.Modem) {
	status := page.MustElement(`#link_status`).MustText()
	stats.Status = status == "Up"

	ds_current_rate := page.MustElement(`#ds_current_rate`).MustText()
	stats.CurrentDown = atoi(strings.TrimSuffix(ds_current_rate, " Kbps"))

	us_current_rate := page.MustElement(`#us_current_rate`).MustText()
	stats.CurrentUp = atoi(strings.TrimSuffix(us_current_rate, " Kbps"))

	ds_maximum_rate := page.MustElement(`#ds_maximum_rate`).MustText()
	stats.MaxDown = atoi(strings.TrimSuffix(ds_maximum_rate, " Kbps"))

	us_maximum_rate := page.MustElement(`#us_maximum_rate`).MustText()
	stats.MaxUp = atoi(strings.TrimSuffix(us_maximum_rate, " Kbps"))

	ds_snr := page.MustElement(`#ds_noise_margin`).MustText()
	stats.SNRDown = atof(strings.TrimSuffix(ds_snr, " dB"))

	us_snr := page.MustElement(`#us_noise_margin`).MustText()
	stats.SNRUp = atof(strings.TrimSuffix(us_snr, " dB"))

	ds_crc := page.MustElement(`#ds_crc`).MustText()
	stats.CRCDown = atoi(ds_crc)

	us_crc := page.MustElement(`#us_crc`).MustText()
	stats.CRCUp = atoi(us_crc)

	ds_fec := page.MustElement(`#ds_fec`).MustText()
	stats.FECDown = atoi(ds_fec)

	us_fec := page.MustElement(`#us_fec`).MustText()
	stats.FECUp = atoi(us_fec)

	stats.DataDown = 0
	stats.DataUp = 0
}

func screenshot(page *rod.Page, pwd string) error {
	buf, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format:  proto.PageCaptureScreenshotFormatJpeg,
		Quality: gson.Int(90),
	})
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(pwd, buf, 0o644)
	if err != nil {
		return err
	}

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

	modem := Cfg.Modem["SpeedportPlus2"]
	var s models.Modem
	browser := rod.New().MustConnect().Trace(false).Timeout(30 * time.Second)
	defer browser.MustClose()

	page := browser.MustPage("http://192.168.2.254")
	page.MustElement(`#userName`).MustWaitVisible()
	page.MustElement(`#userName`).MustInput("admin")
	page.MustElement(`div.row:nth-child(4) > div:nth-child(1) > input:nth-child(1)`).MustInput("bd8rne2b")
	page.MustElement(`.button`).MustClick()
	page.MustElement(`li.main-menu:nth-child(6) > a:nth-child(1) > span:nth-child(1)`).MustWaitVisible()

	dsl := page.MustNavigate(`http://192.168.2.254/status-and-support.html#sub=1&subSub=66`)
	dsl.MustWaitStable()
	parseDSL(dsl, &s)
	// screenshot(dsl, "scr.png")
	dsl.MustElement(`#\33  > a:nth-child(1)`).MustClick()
	dsl.MustWaitStable()
	// screenshot(dsl, "scr2.png")
	parseVoice(dsl, &s)

	s.Host = modem.Host
	err = saveStats(&s, modem.Modem)
	if err != nil {
		log.Fatal(err)
	}
}

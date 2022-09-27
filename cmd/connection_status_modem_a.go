package main

import (
	"context"
	"github.com/DictumMortuum/servus/pkg/models"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/ziutek/telnet"
	"log"
	"regexp"
	"strings"
)

var (
	Cfg         Config
	re_max      = regexp.MustCompile(`Max:\s+Upstream rate = (\d+) Kbps, Downstream rate = (\d+) Kbps`)
	re_cur      = regexp.MustCompile(`Path:\s+\d+, Upstream rate = (\d+) Kbps, Downstream rate = (\d+) Kbps`)
	re_fec_down = regexp.MustCompile(`\nFECErrors:\s+(\d+)`)
	re_fec_up   = regexp.MustCompile(`ATUCFECErrors:\s+(\d+)`)
	re_crc_down = regexp.MustCompile(`\nCRCErrors:\s+(\d+)`)
	re_crc_up   = regexp.MustCompile(`ATUCCRCErrors:\s+(\d+)`)
	re_bytes    = regexp.MustCompile(`bytessent\s+= (\d+)\s+,bytesreceived\s+= (\d+)`)
	re_snr      = regexp.MustCompile(`display dsl snr up=([\d\.]+) down=([\d\.]+) success`)
	re_voip     = regexp.MustCompile(`Status\s+:Enable`)
)

func parseStats(host, user, password, voip string) (*models.Modem, error) {
	var stats models.Modem

	t, err := telnet.Dial("tcp", host)
	if err != nil {
		return nil, err
	}
	defer t.Close()

	t.SetUnixWriteMode(true)
	var data []byte

	err = expect(t, "Login:")
	if err != nil {
		return nil, err
	}

	err = sendln(t, user)
	if err != nil {
		return nil, err
	}

	err = expect(t, "Password:")
	if err != nil {
		return nil, err
	}

	err = sendln(t, password)
	if err != nil {
		return nil, err
	}

	err = expect(t, "WAP>")
	if err != nil {
		return nil, err
	}

	err = sendln(t, "display xdsl connection status")
	if err != nil {
		return nil, err
	}

	data, err = t.ReadBytes('>')
	if err != nil {
		return nil, err
	}

	raw := string(data)

	// TODO: need to parse On Line: 0 Days 3 Hour 17 Min 24 Sec to unix timestamp
	stats.Uptime = 0
	stats.Status = strings.Contains(raw, "Status: Up")

	refs := re_max.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.MaxUp = atoi(match[1])
		stats.MaxDown = atoi(match[2])
	}

	refs = re_cur.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.CurrentUp = atoi(match[1])
		stats.CurrentDown = atoi(match[2])
	}

	refs = re_crc_down.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.CRCDown = atoi(match[1])
	}

	refs = re_crc_up.FindAllStringSubmatch(raw, 1)
	if len(refs) > 0 {
		match := refs[0]
		stats.CRCUp = atoi(match[1])
	}

	refs = re_fec_down.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.FECDown = atoi(match[1])
	}

	refs = re_fec_up.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.FECUp = atoi(match[1])
	}

	err = sendln(t, "display xdsl statistics")
	if err != nil {
		return nil, err
	}

	data, err = t.ReadBytes('>')
	if err != nil {
		return nil, err
	}

	raw = string(data)

	refs = re_bytes.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.DataUp = atoi64(match[1])
		stats.DataDown = atoi64(match[2])
	}

	err = sendln(t, "display dsl snr")
	if err != nil {
		return nil, err
	}

	data, err = t.ReadBytes('>')
	if err != nil {
		return nil, err
	}

	raw = string(data)

	refs = re_snr.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.SNRUp = atof(match[1])
		stats.SNRDown = atof(match[2])
	}

	err = sendln(t, "display waninfo interface "+voip)
	if err != nil {
		return nil, err
	}

	data, err = t.ReadBytes('>')
	if err != nil {
		return nil, err
	}

	raw = string(data)

	refs = re_voip.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		stats.VoipStatus = true
	}

	return &stats, nil
}

func main() {
	loader := confita.NewLoader(
		file.NewBackend("/etc/conf.d/servusrc.yml"),
	)

	err := loader.Load(context.Background(), &Cfg)
	if err != nil {
		log.Fatal(err)
	}

	modem := Cfg.Modem["DG8245V-10"]

	s, err := parseStats(modem.Host+":23", modem.User, modem.Pass, modem.Voip)
	if err != nil {
		log.Fatal(err)
	}

	s.Host = modem.Host
	err = saveStats(s, modem.Modem)
	if err != nil {
		log.Fatal(err)
	}
}

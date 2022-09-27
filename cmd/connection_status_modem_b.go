package main

import (
	"context"
	"github.com/DictumMortuum/servus/pkg/models"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/ziutek/telnet"
	"log"
	"regexp"
)

var (
	Cfg           Config
	re_status     = regexp.MustCompile(`Status\s+: Up`)
	re_max        = regexp.MustCompile(`Max Rate\(Kbps\)    : (\d+)\s+(\d+)`)
	re_cur        = regexp.MustCompile(`Current Rate\(Kbps\): (\d+)\s+(\d+)`)
	re_fec_down   = regexp.MustCompile(`FEC Errors[ :]+(\d+)`)
	re_fec_up     = regexp.MustCompile(`ATU CFEC Errors[ :]+(\d+)`)
	re_crc_down   = regexp.MustCompile(`CRC Errors[ :]+(\d+)`)
	re_crc_up     = regexp.MustCompile(`ATU CCRC Errors[ :]+(\d+)`)
	re_bytes_down = regexp.MustCompile(`Receive Blocks[ :]+(\d+)`)
	re_bytes_up   = regexp.MustCompile(`Transmit Blocks[ :]+(\d+)`)
	re_snr        = regexp.MustCompile(`Noise Margin\(dB\)[ :]+([\d\.]+)\s+([\d\.]+)`)
)

func getStats(host, user, password string) (string, error) {
	t, err := telnet.Dial("tcp", host)
	if err != nil {
		return "", err
	}
	defer t.Close()

	t.SetUnixWriteMode(true)
	var data []byte

	err = expect(t, "ADSL2PlusRouter login:")
	if err != nil {
		return "", err
	}

	err = sendln(t, user)
	if err != nil {
		return "", err
	}

	err = expect(t, "Password:")
	if err != nil {
		return "", err
	}

	err = sendln(t, password)
	if err != nil {
		return "", err
	}

	err = expect(t, "> ")
	if err != nil {
		return "", err
	}

	err = sendSlowly(t, "adsl stats\n")
	if err != nil {
		return "", err
	}

	data, err = t.ReadBytes('>')
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func parseStats(raw string) *models.Modem {
	var stats models.Modem

	refs := re_status.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		stats.Status = true
	}

	refs = re_max.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.MaxDown = atoi(match[1])
		stats.MaxUp = atoi(match[2])
	}

	refs = re_cur.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.CurrentDown = atoi(match[1])
		stats.CurrentUp = atoi(match[2])
	}

	refs = re_snr.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.SNRDown = atof(match[1])
		stats.SNRUp = atof(match[2])
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

	refs = re_bytes_down.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.DataDown = atoi64(match[1])
	}

	refs = re_bytes_up.FindAllStringSubmatch(raw, -1)
	if len(refs) > 0 {
		match := refs[0]
		stats.DataUp = atoi64(match[1])
	}

	return &stats
}

func main() {
	loader := confita.NewLoader(
		file.NewBackend("/etc/conf.d/servusrc.yml"),
	)

	err := loader.Load(context.Background(), &Cfg)
	if err != nil {
		log.Fatal(err)
	}

	modem := Cfg.Modem["TD5130"]

	raw, err := getStats(modem.Host+":23", modem.User, modem.Pass)
	if err != nil {
		log.Fatal(err)
	}

	s := parseStats(raw)
	s.Host = modem.Host
	err = saveStats(s, modem.Modem)
	if err != nil {
		log.Fatal(err)
	}
}

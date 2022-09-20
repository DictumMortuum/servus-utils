package main

import (
	"context"
	"encoding/json"
	"github.com/DictumMortuum/servus/pkg/models"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/jmoiron/sqlx"
	"github.com/ziutek/telnet"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Databases map[string]string `config:"databases"`
	Host      string            `config:"host"`
	Modem     string            `config:"modem"`
	User      string            `config:"user"`
	Pass      string            `config:"pass"`
}

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

const timeout = 20 * time.Second

func expect(t *telnet.Conn, d ...string) error {
	err := t.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}

	err = t.SkipUntil(d...)
	if err != nil {
		return err
	}

	return nil
}

func sendln(t *telnet.Conn, s string) error {
	err := t.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}

	buf := make([]byte, len(s)+1)
	copy(buf, s)
	buf[len(s)] = '\n'

	_, err = t.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func sendSlowly(t *telnet.Conn, s string) error {
	err := t.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}

	for _, c := range s {
		_, err = t.Write([]byte(string(c)))
		if err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func getStats(host, user, password string) (string, error) {
	t, err := telnet.Dial("tcp", host)
	if err != nil {
		return "", err
	}

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

func atoi(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}

func atoi64(s string) int64 {
	i, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 32)
	return i
}

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
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

func saveStats(s *models.Modem) error {
	payload, err := json.Marshal(s)
	if err != nil {
		return err
	}

	db, err := sqlx.Connect("mysql", Cfg.Databases["mariadb"])
	if err != nil {
		return err
	}
	defer db.Close()

	q := `update tkeyval set json = :json, date = NOW() where id = :id`
	_, err = db.NamedExec(q, map[string]any{
		"id":   "TD5130",
		"json": string(payload),
	})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	loader := confita.NewLoader(
		file.NewBackend("/etc/conf.d/servusrc.yml"),
		file.NewBackend("/etc/conf.d/modem_b.yml"),
	)

	err := loader.Load(context.Background(), &Cfg)
	if err != nil {
		log.Fatal(err)
	}

	raw, err := getStats(Cfg.Host+":23", Cfg.User, Cfg.Pass)
	if err != nil {
		log.Fatal(err)
	}

	s := parseStats(raw)
	err = saveStats(s)
	if err != nil {
		log.Fatal(err)
	}
}

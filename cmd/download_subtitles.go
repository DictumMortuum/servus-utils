package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	rofi "github.com/DictumMortuum/gofi"
	"github.com/golang-jwt/jwt"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/heetch/confita/backend/flags"
	"github.com/odwrtw/opensubtitles"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultUserAgent = "osdb-go"
	defaultEndpoint  = "https://api.opensubtitles.com"
)

type Config struct {
	Filename string `config:"filename"`
	Lang     string `config:"lang"`
	ID       string `config:"id"`
	User     string `config:"opensubtitles_user"`
	Pass     string `config:"opensubtitles_pass"`
	Key      string `config:"opensubtitles_key"`
}

func Download(cfg Config, id int) (map[string]any, error) {
	type Payload struct {
		FileID int `json:"file_id"`
	}

	data := Payload{
		FileID: id,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", defaultEndpoint+"/api/v1/download", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Api-Key", cfg.Key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rs map[string]any
	err = json.Unmarshal(raw, &rs)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func fileNameWithoutExtSliceNotation(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func srtForVideo(fileName string) string {
	return fileNameWithoutExtSliceNotation(fileName) + ".srt"
}

func Download2(rs map[string]any, filename string) error {
	var url string
	if val, ok := rs["link"]; ok {
		url = val.(string)
	} else {
		log.Fatal("Could not find 'link' attribute.")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(srtForVideo(filename))
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	loader := confita.NewLoader(
		flags.NewBackend(),
		file.NewBackend("/etc/servus.yml"),
	)

	cfg := Config{
		Lang: "en",
	}

	loader.Load(context.Background(), &cfg)

	if cfg.Filename != "" {
		c := NewClient(cfg.Key, cfg.User, cfg.Pass)
		res, err := c.SearchByFile(cfg.Filename, []string{"en"})
		if err != nil {
			log.Fatal(err)
		}

		payload := map[string]any{}
		for _, sub := range res {
			for _, f := range sub.Subtitle.Files {
				payload[f.FileName] = f.ID
			}
		}

		opts := rofi.GofiOptions{
			Description: "subtitles",
		}

		options, err := rofi.FromInterface(&opts, payload)
		if err != nil {
			log.Fatal(err)
		}

		key := payload[options[0]].(int)
		rs, err := Download(cfg, key)
		if err != nil {
			log.Fatal(err)
		}

		if val, ok := rs["message"]; ok {
			if count, ok := rs["remaining"]; ok {
				fmt.Println(val, "remaining:", count)
			}
		}

		err = Download2(rs, cfg.Filename)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type Client struct {
	UserAgent string
	Endpoint  string
	APIKey    string
	Username  string
	Password  string
	Token     *jwt.Token
	User      *opensubtitles.User
}

// NewClient returns a new client
func NewClient(apiKey, username, password string) *Client {
	return &Client{
		Endpoint: defaultEndpoint,
		APIKey:   apiKey,
		Username: username,
		Password: password,
	}
}

func (c *Client) get(url string, resp interface{}, auth bool) error {
	return c.request("GET", url, nil, resp, auth)
}

func (c *Client) post(url string, data, resp interface{}, auth bool) error {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(data); err != nil {
		return err
	}

	return c.request("POST", url, body, resp, auth)
}

func (c *Client) request(method, url string, body io.Reader, respData interface{}, auth bool) error {
	url = c.Endpoint + url

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Api-Key", c.APIKey)

	if auth {
		// No token or token expired
		if c.Token == nil || c.Token.Claims.Valid() != nil {
			if _, err := c.Login(); err != nil {
				return err
			}
		}

		req.Header.Add("Authorization", "Bearer "+c.Token.Raw)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"opensubtitles: invalid status code %s (%d)",
			resp.Status, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(respData)
}

func (c *Client) Search(q opensubtitles.SubtitleQueryParameters) ([]*SubtitleData, error) {
	resp := struct {
		TotalPages int             `json:"total_pages"`
		TotalCount int             `json:"total_count"`
		Page       int             `json:"page"`
		Data       []*SubtitleData `json:"data"`
	}{}

	if err := c.get("/api/v1/subtitles?"+q.Encode(), &resp, false); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Client) Login() (*opensubtitles.UserLogin, error) {
	credentials := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: c.Username,
		Password: c.Password,
	}

	login := &opensubtitles.UserLogin{}
	if err := c.post("/api/v1/login", &credentials, &login, false); err != nil {
		return nil, err
	}

	parser := &jwt.Parser{}
	claims := jwt.StandardClaims{}
	token, _, err := parser.ParseUnverified(login.Token, &claims)
	if err != nil {
		return nil, err
	}

	if err := token.Claims.Valid(); err != nil {
		return nil, err
	}

	c.Token = token
	return login, nil
}

func (c *Client) SearchByFile(path string, langs []string) ([]*SubtitleData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash, err := opensubtitles.Hash(file)
	if err != nil {
		return nil, err
	}

	q := opensubtitles.SubtitleQueryParameters{
		Query:          filepath.Base(path),
		MovieHash:      opensubtitles.HashString(hash),
		MovieHashMatch: "only",
		Languages:      langs,
	}

	return c.Search(q)
}

// SubtitleData holds the subtitle
type SubtitleData struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Subtitle Subtitle `json:"attributes"`
}

// Subtitle represents a subtile response
type Subtitle struct {
	ID               string                       `json:"subtitle_id"`
	Language         string                       `json:"language"`
	DownloadCount    int                          `json:"download_count"`
	NewDownloadCount int                          `json:"new_download_count"`
	HearingImpared   bool                         `json:"hearing_impared"`
	HD               bool                         `json:"hd"`
	FPS              float64                      `json:"fps"`
	Votes            int                          `json:"votes"`
	Points           int                          `json:"points"`
	Rating           int                          `json:"rating"`
	FromTrusted      bool                         `json:"from_trusted"`
	ForeignPartsOnly bool                         `json:"foreign_parts_only"`
	AutoTranslation  bool                         `json:"auto_translation"`
	AITranslated     bool                         `json:"ai_translated"`
	UploadDate       time.Time                    `json:"upload_date"`
	Release          string                       `json:"release"`
	Comments         string                       `json:"comments"`
	LegacySubtitleID int                          `json:"legacy_subtitle_id"`
	URL              string                       `json:"url"`
	FeatureDetails   opensubtitles.FeatureDetails `json:"feature_details"`
	Uploader         opensubtitles.Uploader       `json:"uploader"`
	RelatedLinks     []any                        `json:"related_links"`
	Files            []opensubtitles.File         `json:"files"`
	MovieHashMatch   bool                         `json:"movie_hash_match"`

	// Format and MachineTranslated are not listed because they are not
	// required and their type it not defined in the documentation...
}

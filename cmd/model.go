package main

type Config struct {
	Databases map[string]string `config:"databases"`
	Host      string            `config:"host"`
	Modem     string            `config:"modem"`
	User      string            `config:"user"`
	Pass      string            `config:"pass"`
	Voip      string            `config:"voip"`
}

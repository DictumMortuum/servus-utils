package main

type ModemConfig struct {
	Host  string `config:"host"`
	Modem string `config:"modem"`
	User  string `config:"user"`
	Pass  string `config:"pass"`
	Voip  string `config:"voip"`
}

type Config struct {
	Databases map[string]string      `config:"databases"`
	Modem     map[string]ModemConfig `config:"modem"`
}

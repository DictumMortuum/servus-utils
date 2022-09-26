package main

import (
	"encoding/json"
	"github.com/DictumMortuum/servus/pkg/models"
	"github.com/jmoiron/sqlx"
)

func saveStats(s *models.Modem, id string) error {
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
		"id":   id,
		"json": string(payload),
	})
	if err != nil {
		return err
	}

	return nil
}

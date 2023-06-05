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

func saveTragedy(title, section, part, url string) error {
	db, err := sqlx.Connect("mysql", Cfg.Databases["mariadb"])
	if err != nil {
		return err
	}
	defer db.Close()

	q := `insert into ttragedy (title, section, part, url) values (:title, :section, :part, :url) on duplicate key update id = id`
	_, err = db.NamedExec(q, map[string]any{
		"title":   title,
		"section": section,
		"part":    part,
		"url":     url,
	})
	if err != nil {
		return err
	}

	return nil
}

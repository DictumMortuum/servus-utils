package scrape

import (
	"database/sql"
	// "github.com/DictumMortuum/servus/pkg/models"
	"github.com/jmoiron/sqlx"
)

type Price struct {
	Id   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func GetPrices(db *sqlx.DB) ([]Price, error) {
	var rs []Price
	err := db.Select(&rs, "select id, name from tboardgameprices")
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func GetCachedPrices(db *sqlx.DB) ([]Price, error) {
	var rs []Price
	err := db.Select(&rs, "select id, name from tboardgamepricescached")
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func IgnorePrice(db *sqlx.DB, id int64) error {
	q := `
		update
			tboardgameprices
		set
			ignored = true
		where
			id = ?
	`
	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}

func GetIgnored(db *sqlx.DB) ([]string, error) {
	var rs []string
	err := db.Select(&rs, "select name from tboardgamepricesignored")
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func SetIgnored(db *sqlx.DB, name string) error {
	q := `insert into tboardgamepricesignored (name) values (:name)`
	_, err := db.NamedExec(q, map[string]any{
		"name": name,
	})
	if err != nil {
		return err
	}

	return nil
}

func InsertCachedPrice(db *sqlx.DB, payload map[string]any) error {
	q := `insert into tboardgamepricescached (name,store_id,store_thumb,price,stock,url) values (:name,:store_id,:store_thumb,:price,:stock,:url)`
	_, err := db.NamedExec(q, payload)
	if err != nil {
		return err
	}

	return nil
}

func CleanCachedPrices(db *sqlx.DB, id int) error {
	q := `delete from tboardgamepricescached where store_id = ?`
	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}

func DeleteCachedPrice(db *sqlx.DB, id int64) error {
	q := `delete from tboardgamepricescached where id = ?`
	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}

func PriceExists(db *sqlx.DB, payload map[string]any) (*sql.NullInt64, error) {
	var id sql.NullInt64

	q := `select id from tboardgameprices where store_id = :store_id and name = :name and boardgame_id is not null`
	stmt, err := db.PrepareNamed(q)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.Get(&id, payload)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	_, err = db.NamedExec(q, payload)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func UpdatePrice(db *sqlx.DB, id *sql.NullInt64, payload map[string]any) error {
	payload["id"] = id

	q := `
		update
			tboardgameprices
		set
			store_thumb = :store_thumb,
			stock = :stock,
			price = :price,
			url = :url,
			batch = 1
		where
			id = :id
	`
	_, err := db.NamedExec(q, payload)
	if err != nil {
		return err
	}

	return nil
}

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
	q := `
		insert into tboardgamepricescached (
			name,
			store_id,
			store_thumb,
			price,
			stock,
			url
		) values (
			:name,
			:store_id,
			:store_thumb,
			:price,
			:stock,
			:url
		) on duplicate key update id = id
	`
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

func PriceExists(db *sqlx.DB, payload map[string]any) (*sql.NullInt64, *sql.NullInt64, error) {
	type result struct {
		Id          sql.NullInt64 `db:"id"`
		BoardgameId sql.NullInt64 `db:"boardgame_id"`
	}

	var rs result

	q := `select id, boardgame_id from tboardgameprices where store_id = :store_id and name = :name and boardgame_id is not null`
	stmt, err := db.PrepareNamed(q)
	if err != nil {
		return nil, nil, err
	}
	defer stmt.Close()

	err = stmt.Get(&rs, payload)
	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	_, err = db.NamedExec(q, payload)
	if err != nil {
		return nil, nil, err
	}

	return &rs.Id, &rs.BoardgameId, nil
}

func UpdatePrice(db *sqlx.DB, payload map[string]any) error {
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

func InsertMapping(db *sqlx.DB, payload map[string]any) (bool, error) {
	q := `
		insert into tboardgamepricesmap (
			boardgame_id,
			name
		) values (
			:boardgame_id,
			:name
		) on duplicate key update id = id
	`

	rs, err := db.NamedExec(q, payload)
	if err != nil {
		return false, err
	}

	rows, err := rs.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

func InsertHistories(db *sqlx.DB, payload map[string]any) (bool, error) {
	q := `
		insert into	tboardgamepriceshistory (
			boardgame_id,
			cr_date,
			price,
			stock,
			store_id
		) values (
			:boardgame_id,
			date_add(date_add(LAST_DAY(NOW()), interval 1 day), interval -1 month),
			:price,
			:stock,
			:store_id
		) on duplicate key update id = id
	`

	rs, err := db.NamedExec(q, payload)
	if err != nil {
		return false, err
	}

	rows, err := rs.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

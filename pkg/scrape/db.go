package scrape

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"strings"
)

type Price struct {
	Id          int64   `db:"id" json:"id"`
	StoreId     int64   `db:"store_id" json:"store_id"`
	BoardgameId int64   `db:"boardgame_id" json:"boardgame_id"`
	Name        string  `db:"name" json:"name"`
	Price       float64 `db:"price" json:"price"`
	Stock       int     `db:"stock" json:"stock"`
}

type Ignored struct {
	Name    string `db:"name"`
	StoreId int64  `db:"store_id"`
}

func GetPrices(db *sqlx.DB) ([]Price, error) {
	var rs []Price
	err := db.Select(&rs, "select id, store_id, name from tboardgameprices")
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func GetNewlyMappedPrices(db *sqlx.DB) ([]Price, error) {
	var rs []Price
	err := db.Select(&rs, "select id, boardgame_id, store_id, price, stock, name from tboardgameprices where mapped = 0 and boardgame_id is not null")
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func GetCachedPrices(db *sqlx.DB) ([]Price, error) {
	var rs []Price
	err := db.Select(&rs, "select id, store_id, name from tboardgamepricescached")
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

func GetIgnored(db *sqlx.DB) ([]Ignored, error) {
	var rs []Ignored
	err := db.Select(&rs, "select name, store_id from tboardgamepricesignored")
	if err != nil {
		return nil, err
	}

	retval := []Ignored{}
	for _, item := range rs {
		check := strings.ToLower(item.Name)
		check = removeAccents(check)
		retval = append(retval, Ignored{
			Name:    check,
			StoreId: item.StoreId,
		})
	}

	return retval, nil
}

func GetIgnoredNames(db *sqlx.DB) ([]string, error) {
	var rs []string
	err := db.Select(&rs, "select name from tboardgamenamesignored")
	if err != nil {
		return nil, err
	}

	retval := []string{}
	for _, item := range rs {
		check := strings.ToLower(item)
		check = removeAccents(check)
		retval = append(retval, check)
	}

	return retval, nil
}

// func SetIgnored(db *sqlx.DB, name string) error {
// 	q := `insert into tboardgamepricesignored (name) values (:name)`
// 	_, err := db.NamedExec(q, map[string]any{
// 		"name": name,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func CreatePricesFromCachedPrices(db *sqlx.DB) (bool, error) {
	q := `
		insert into tboardgameprices (
			cr_date,
			boardgame_id,
			store_id,
			store_thumb,
			name,
			price,
			stock,
			url,
			batch,
			mapped,
			ignored
		)
		select
			NOW(),
			NULL,
			store_id,
			store_thumb,
			name,
			price,
			stock,
			url,
			1,
			false,
			false
		from
			tboardgamepricescached
		on duplicate key update cr_date = NOW()
	`

	rs, err := db.Exec(q)
	if err != nil {
		return false, err
	}

	rows, err := rs.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

func InsertCachedPrice(db *sqlx.DB, payload map[string]any) (bool, error) {
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

func CleanCachedPrices(db *sqlx.DB, id int) error {
	q := `delete from tboardgamepricescached where store_id = ?`
	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}

func DeleteCachedPrices(db *sqlx.DB) error {
	q := `delete from tboardgamepricescached`
	_, err := db.Exec(q)
	if err != nil {
		return err
	}

	q = `delete from tboardgameprices where boardgame_id is null and mapped = 0`
	_, err = db.Exec(q)
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

func MapPrice(db *sqlx.DB, payload any) error {
	q := `
		update
			tboardgameprices
		set
			batch = 1,
			mapped = 1
		where
			id = :id
	`
	_, err := db.NamedExec(q, payload)
	if err != nil {
		return err
	}

	return nil
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
			batch = 1,
			mapped = 1
		where
			id = :id and
			boardgame_id is not null
	`
	_, err := db.NamedExec(q, payload)
	if err != nil {
		return err
	}

	return nil
}

func InsertMapping(db *sqlx.DB, payload any) (bool, error) {
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

func InsertHistories(db *sqlx.DB, payload any) (bool, error) {
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

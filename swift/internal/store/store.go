package store

import (
	"database/sql"
	"github.com/mbartnicki80/swift/internal/parser"
)

func InsertRowsToDatabase(db *sql.DB, records []parser.SwiftRecord) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO swift_codes (country_iso2_code, swift_code,
		                         code_type, bank_name,
		                         address, town_name,
		                         country_name, time_zone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (swift_code) DO NOTHING
		`

	for _, record := range records {
		_, err := tx.Exec(query, record.ISO2, record.Code, record.Type,
			record.Name, record.Address, record.Town, record.Country, record.TimeZone)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

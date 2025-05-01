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

	insertSwift := `
		INSERT INTO swift_codes (country_iso2_code, swift_code,
		                         bank_name, address, town_name,
		                         country_name, is_headquarter)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (swift_code) DO NOTHING
	`

	insertBranch := `
		INSERT INTO branches (swift_code, headquarter)
		VALUES ($1, $2)
		ON CONFLICT (swift_code) DO NOTHING
	`

	for _, record := range records {
		_, err := tx.Exec(insertSwift, record.ISO2Code, record.SwiftCode, record.BankName,
			record.Address, record.Town, record.Country, record.IsHeadquarter)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}

		if !record.IsHeadquarter {
			hqCode := record.SwiftCode[:8] + "XXX"

			_, err := tx.Exec(insertSwift, record.ISO2Code, hqCode, record.BankName,
				record.Address, record.Town, record.Country, true)
			if err != nil {
				err := tx.Rollback()
				if err != nil {
					return err
				}
				return err
			}

			_, err = tx.Exec(insertBranch, record.SwiftCode, hqCode)
			if err != nil {
				err := tx.Rollback()
				if err != nil {
					return err
				}
				return err
			}
		}
	}

	return tx.Commit()
}

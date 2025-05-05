package store

import (
	"database/sql"
	"github.com/mbartnicki80/swift/internal/model"
	"github.com/mbartnicki80/swift/internal/parser"
	"log"
)

func InsertBranches(tx *sql.Tx, records []parser.SwiftRecord) error {
	insertIntoBranchQuery := `
		INSERT INTO branches (swift_code, headquarter)
		VALUES ($1, $2)
		ON CONFLICT (swift_code) DO NOTHING
	`
	var err error
	for _, record := range records {
		if !record.IsHeadquarter {
			hqCode := record.SwiftCode[:8] + "XXX"
			var exists bool
			err = tx.QueryRow("SELECT EXISTS (SELECT 1 FROM swift_codes WHERE swift_code = $1)", hqCode).Scan(&exists)
			if err != nil {
				tx.Rollback()
				return err
			}

			var hqPtr *string
			if !exists {
				hqPtr = nil
			} else {
				hqPtr = &hqCode
			}

			_, err = tx.Exec(insertIntoBranchQuery, record.SwiftCode, hqPtr)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}

func InsertRowsToDatabase(db *sql.DB, records []parser.SwiftRecord) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	insertSwiftCodeQuery := `
		INSERT INTO swift_codes (country_iso2_code, swift_code,
		                         bank_name, address,
		                         country_name, is_headquarter)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (swift_code) DO NOTHING
	`

	for _, record := range records {
		_, err := tx.Exec(insertSwiftCodeQuery, record.ISO2Code, record.SwiftCode, record.BankName,
			record.Address, record.Country, record.IsHeadquarter)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return InsertBranches(tx, records)
}

func FetchSwiftCode(db *sql.DB, swiftCode string) (model.SwiftCode, []model.SwiftCode, error) {
	FetchSwiftCodeQuery := `
	SELECT swift_code, address, country_name, is_headquarter, country_iso2_code, bank_name
	FROM swift_codes
	WHERE swift_code = $1
	`
	FetchBranchQuery := `
	SELECT swift_codes.swift_code, swift_codes.address, swift_codes.is_headquarter,
		   swift_codes.country_iso2_code, swift_codes.bank_name
	FROM branches
	JOIN swift_codes ON swift_codes.swift_code = branches.swift_code
	WHERE branches.headquarter = $1
	`

	res, err := db.Query(FetchSwiftCodeQuery, swiftCode)
	if err != nil {
		return model.SwiftCode{}, nil, err
	}
	defer func(res *sql.Rows) {
		err := res.Close()
		if err != nil {
			return
		}
	}(res)

	if !res.Next() {
		return model.SwiftCode{}, nil, sql.ErrNoRows
	}

	var result model.SwiftCode
	err = res.Scan(&result.SwiftCode, &result.Address, &result.CountryName, &result.IsHeadquarter, &result.CountryISO2, &result.BankName)
	if err != nil {
		return model.SwiftCode{}, nil, err
	}

	if !result.IsHeadquarter {
		return result, nil, nil
	}

	rows, err := db.Query(FetchBranchQuery, swiftCode)
	if err != nil {
		return result, nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	var branches []model.SwiftCode
	for rows.Next() {
		var branch model.SwiftCode
		err := rows.Scan(&branch.SwiftCode, &branch.Address, &branch.IsHeadquarter, &branch.CountryISO2, &branch.BankName)
		if err != nil {
			return result, nil, err
		}
		branches = append(branches, branch)
	}

	if err := rows.Err(); err != nil {
		return result, nil, err
	}

	return result, branches, nil
}

func FetchSwiftCodesByCountry(db *sql.DB, countryCode string) ([]model.SwiftCode, error) {
	FetchSwiftCodesByCountryQuery := `
		SELECT address, bank_name, country_iso2_code, is_headquarter, swift_code, country_name
		FROM swift_codes
		WHERE country_iso2_code = $1
		`
	rows, err := db.Query(FetchSwiftCodesByCountryQuery, countryCode)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	var results []model.SwiftCode
	for rows.Next() {
		var result model.SwiftCode
		err := rows.Scan(&result.Address, &result.BankName, &result.CountryISO2, &result.IsHeadquarter, &result.SwiftCode, &result.CountryName)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func InsertNewSwiftCode(db *sql.DB, swiftCode model.SwiftCode) error {
	insertSwiftCodeQuery := `
		INSERT INTO swift_codes (country_iso2_code, swift_code,
		                         bank_name, address, country_name, 
		                         is_headquarter)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (swift_code) DO NOTHING
	`

	_, err := db.Exec(insertSwiftCodeQuery, swiftCode.CountryISO2, swiftCode.SwiftCode, swiftCode.BankName,
		swiftCode.Address, swiftCode.CountryName, swiftCode.IsHeadquarter)
	if err != nil {
		return err
	}

	if !swiftCode.IsHeadquarter {
		hq := swiftCode.SwiftCode[:8] + "XXX"
		var exists bool
		err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM swift_codes WHERE swift_code = $1)", hq).Scan(&exists)
		if err != nil {
			return err
		}
		var hqPtr *string
		if !exists {
			hqPtr = nil
		} else {
			hqPtr = &hq
		}

		insertIntoBranchQuery := `
		INSERT INTO branches (swift_code, headquarter)
		VALUES ($1, $2)
		ON CONFLICT (swift_code) DO NOTHING
		`
		_, err = db.Exec(insertIntoBranchQuery, swiftCode.SwiftCode, hqPtr)
		if err != nil {
			return err
		}
	} else {
		updateBranchesQuery := `
			UPDATE branches 
			SET headquarter = $1
			WHERE headquarter IS NULL AND swift_code LIKE $2
		`
		_, err = db.Exec(updateBranchesQuery, swiftCode.SwiftCode, swiftCode.SwiftCode[:8]+"%")
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteSwiftCode(db *sql.DB, swiftCode string) error {
	deleteQuery := "DELETE FROM swift_codes WHERE swift_code=$1"
	_, err := db.Exec(deleteQuery, swiftCode)
	if err != nil {
		return err
	}
	return nil
}

package store

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mbartnicki80/swift/internal/model"
	"github.com/mbartnicki80/swift/internal/parser"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	insertSwiftCodeQuery = `
INSERT INTO swift_codes (country_iso2_code, swift_code,
                         bank_name, address,
                         country_name, is_headquarter)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (swift_code) DO NOTHING
`

	insertBranchQuery = `
INSERT INTO branches (swift_code, headquarter)
VALUES ($1, $2)
ON CONFLICT (swift_code) DO NOTHING
`

	deleteSwiftCodeQuery = `
DELETE FROM swift_codes WHERE swift_code=$1
`

	fetchSwiftCodesByCountryQuery = `
SELECT address, bank_name, country_iso2_code, is_headquarter, swift_code, country_name
FROM swift_codes
WHERE country_iso2_code = $1
`

	fetchSwiftCodeQuery = `
SELECT swift_code, address, country_name, is_headquarter, country_iso2_code, bank_name
FROM swift_codes
WHERE swift_code = $1
`

	fetchBranchQuery = `
SELECT swift_codes.swift_code, swift_codes.address, swift_codes.is_headquarter,
	   swift_codes.country_iso2_code, swift_codes.bank_name
FROM branches
JOIN swift_codes ON swift_codes.swift_code = branches.swift_code
WHERE branches.swift_code = $1
`

	selectExists = `
SELECT EXISTS (SELECT 1 FROM swift_codes WHERE swift_code = $1)
`
	updateBranchesQuery = `
			UPDATE branches 
			SET headquarter = $1
			WHERE headquarter IS NULL AND swift_code LIKE $2
		`
)

func TestInsertRowsToDatabase(t *testing.T) {
	t.Run("successful insertion", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		records := []parser.SwiftRecord{
			{
				ISO2Code:      "PL",
				SwiftCode:     "PKOPPLPW002",
				BankName:      "PKO",
				Address:       "Krakow",
				Country:       "Poland",
				IsHeadquarter: false,
			},
			{
				ISO2Code:      "PL",
				SwiftCode:     "PKOPPLPWXXX",
				BankName:      "PKO",
				Address:       "Warsaw",
				Country:       "Poland",
				IsHeadquarter: true,
			},
		}

		mock.ExpectBegin()

		for _, r := range records {
			mock.ExpectExec(insertSwiftCodeQuery).
				WithArgs(r.ISO2Code, r.SwiftCode, r.BankName, r.Address, r.Country, r.IsHeadquarter).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		mock.ExpectQuery(selectExists).
			WithArgs("PKOPPLPWXXX").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		mock.ExpectExec(insertBranchQuery).
			WithArgs("PKOPPLPW002", "PKOPPLPWXXX").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err = InsertRowsToDatabase(db, records)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("begin fails", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

		err = InsertRowsToDatabase(db, nil)
		require.ErrorContains(t, err, "begin failed")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("insert branch without headquarter success", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		r := parser.SwiftRecord{
			ISO2Code:      "PL",
			SwiftCode:     "PKOPPLPW002",
			BankName:      "PKO",
			Address:       "Krakow",
			Country:       "Poland",
			IsHeadquarter: false,
		}

		mock.ExpectBegin()
		mock.ExpectExec(insertSwiftCodeQuery).
			WithArgs(r.ISO2Code, r.SwiftCode, r.BankName, r.Address, r.Country, r.IsHeadquarter).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(selectExists).
			WithArgs("PKOPPLPWXXX").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectExec(insertBranchQuery).
			WithArgs(r.SwiftCode, nil).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = InsertRowsToDatabase(db, []parser.SwiftRecord{r})
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFetchSwiftCode(t *testing.T) {
	t.Run("returns headquarter and branches", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		hqCode := "PKOPPLPWXXX"

		mock.ExpectQuery(fetchSwiftCodeQuery).WithArgs(hqCode).
			WillReturnRows(sqlmock.NewRows([]string{
				"swift_code", "address", "country_name", "is_headquarter", "country_iso2_code", "bank_name",
			}).AddRow(hqCode, "Warsaw", "Poland", true, "PL", "PKO"))

		mock.ExpectQuery(fetchBranchQuery).WithArgs(hqCode).
			WillReturnRows(sqlmock.NewRows([]string{
				"swift_code", "address", "is_headquarter", "country_iso2_code", "bank_name",
			}).AddRow("PKOPPLPW002", "Krakow", false, "PL", "PKO"))

		hq, branches, err := FetchSwiftCode(db, hqCode)
		require.NoError(t, err)
		require.Equal(t, hq.SwiftCode, hqCode)
		require.Len(t, branches, 1)
		require.Equal(t, branches[0].SwiftCode, "PKOPPLPW002")
	})

	t.Run("returns single branch without branches", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		swift := "PKOPPLPW002"
		mock.ExpectQuery(fetchSwiftCodeQuery).
			WithArgs(swift).
			WillReturnRows(sqlmock.NewRows([]string{
				"swift_code", "address", "country_name", "is_headquarter", "country_iso2_code", "bank_name",
			}).AddRow(swift, "Krakow", "Poland", false, "PL", "PKO"))

		hq, branches, err := FetchSwiftCode(db, swift)
		require.NoError(t, err)
		require.Equal(t, hq.SwiftCode, swift)
		require.Nil(t, branches)
	})

	t.Run("no rows found", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(fetchSwiftCodeQuery).
			WithArgs("UNKNOWN").
			WillReturnRows(sqlmock.NewRows([]string{}))

		_, _, err = FetchSwiftCode(db, "UNKNOWN")
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("db error on swift fetch", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(fetchSwiftCodeQuery).
			WithArgs("ERR").
			WillReturnError(errors.New("db error"))

		_, _, err = FetchSwiftCode(db, "ERR")
		require.ErrorContains(t, err, "db error")
	})
}

func TestFetchSwiftCodesByCountry(t *testing.T) {
	t.Run("returns swift codes for country", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		country := "PL"
		mock.ExpectQuery(fetchSwiftCodesByCountryQuery).
			WithArgs(country).
			WillReturnRows(sqlmock.NewRows([]string{
				"address", "bank_name", "country_iso2_code", "is_headquarter", "swift_code", "country_name",
			}).AddRow("Warsaw", "PKO", "PL", true, "PKOPPLPWXXX", "Poland").
				AddRow("Krakow", "PKO", "PL", false, "PKOPPLPW002", "Poland"))

		result, err := FetchSwiftCodesByCountry(db, country)
		require.NoError(t, err)
		require.Len(t, result, 2)
	})

	t.Run("no results found", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(fetchSwiftCodesByCountryQuery).
			WithArgs("XX").
			WillReturnRows(sqlmock.NewRows([]string{}))

		result, err := FetchSwiftCodesByCountry(db, "XX")
		require.NoError(t, err)
		require.Empty(t, result)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(fetchSwiftCodesByCountryQuery).
			WithArgs("ERR").
			WillReturnError(errors.New("db error"))

		_, err = FetchSwiftCodesByCountry(db, "ERR")
		require.ErrorContains(t, err, "db error")
	})
}

func TestInsertNewSwiftCode(t *testing.T) {
	t.Run("valid headquarter insert without dangling branches", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		swift := model.SwiftCode{
			CountryISO2:   "PL",
			SwiftCode:     "PKOPPLPWXXX",
			BankName:      "PKO",
			Address:       "Warsaw",
			CountryName:   "Poland",
			IsHeadquarter: true,
		}

		mock.ExpectExec(insertSwiftCodeQuery).
			WithArgs(
				swift.CountryISO2,
				swift.SwiftCode,
				swift.BankName,
				swift.Address,
				swift.CountryName,
				swift.IsHeadquarter,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(updateBranchesQuery).
			WithArgs(swift.SwiftCode, swift.SwiftCode[:8]+"%").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err = InsertNewSwiftCode(db, swift)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("valid headquarter insert with dangling branch", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		swift := model.SwiftCode{
			CountryISO2:   "PL",
			SwiftCode:     "PKOPPLPWXXX",
			BankName:      "PKO",
			Address:       "Warsaw",
			CountryName:   "Poland",
			IsHeadquarter: true,
		}

		mock.ExpectExec(insertSwiftCodeQuery).
			WithArgs(
				swift.CountryISO2,
				swift.SwiftCode,
				swift.BankName,
				swift.Address,
				swift.CountryName,
				swift.IsHeadquarter,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(updateBranchesQuery).
			WithArgs(swift.SwiftCode, swift.SwiftCode[:8]+"%").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = InsertNewSwiftCode(db, swift)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("valid branch insert", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		swift := model.SwiftCode{
			CountryISO2:   "PL",
			SwiftCode:     "PKOPPLPWABC",
			BankName:      "PKO",
			Address:       "Warsaw",
			CountryName:   "Poland",
			IsHeadquarter: false,
		}
		hq := "PKOPPLPWXXX"

		mock.ExpectExec(insertSwiftCodeQuery).
			WithArgs(
				swift.CountryISO2,
				swift.SwiftCode,
				swift.BankName,
				swift.Address,
				swift.CountryName,
				swift.IsHeadquarter).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(selectExists).
			WithArgs(hq).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		mock.ExpectExec(insertBranchQuery).
			WithArgs(swift.SwiftCode, hq).WillReturnResult(sqlmock.NewResult(1, 1))

		err = InsertNewSwiftCode(db, swift)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("branch insert without headquarter", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		swift := model.SwiftCode{
			CountryISO2:   "PL",
			SwiftCode:     "PKOPPLPWABC",
			BankName:      "PKO",
			Address:       "Warsaw",
			CountryName:   "Poland",
			IsHeadquarter: false,
		}
		hq := "PKOPPLPWXXX"

		mock.ExpectExec(insertSwiftCodeQuery).
			WithArgs(
				swift.CountryISO2,
				swift.SwiftCode,
				swift.BankName,
				swift.Address,
				swift.CountryName,
				swift.IsHeadquarter).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(selectExists).
			WithArgs(hq).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectExec(insertBranchQuery).
			WithArgs(swift.SwiftCode, nil).WillReturnResult(sqlmock.NewResult(1, 1))

		err = InsertNewSwiftCode(db, swift)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty swift code", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		swift := model.SwiftCode{
			CountryISO2:   "PL",
			SwiftCode:     "",
			BankName:      "PKO",
			Address:       "Warsaw",
			CountryName:   "Poland",
			IsHeadquarter: true,
		}

		mock.ExpectExec(insertSwiftCodeQuery).
			WithArgs(
				swift.CountryISO2,
				swift.SwiftCode,
				swift.BankName,
				swift.Address,
				swift.CountryName,
				swift.IsHeadquarter,
			).
			WillReturnError(errors.New("invalid swift code"))

		err = InsertNewSwiftCode(db, swift)
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db returns error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()

		swift := model.SwiftCode{
			CountryISO2:   "PL",
			SwiftCode:     "PKOPPLPWXXX",
			BankName:      "PKO",
			Address:       "Warsaw",
			CountryName:   "Poland",
			IsHeadquarter: true,
		}

		mock.ExpectExec(insertSwiftCodeQuery).
			WithArgs(
				swift.CountryISO2,
				swift.SwiftCode,
				swift.BankName,
				swift.Address,
				swift.CountryName,
				swift.IsHeadquarter,
			).
			WillReturnError(errors.New("connection lost"))

		err = InsertNewSwiftCode(db, swift)
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

}

func TestDeleteSwiftCode(t *testing.T) {
	t.Run("valid removal", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		swiftCode := "PKOPPLPWXXX"
		mock.ExpectExec(deleteSwiftCodeQuery).
			WithArgs(swiftCode).WillReturnResult(sqlmock.NewResult(1, 1))
		err = DeleteSwiftCode(db, swiftCode)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty swift code", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.ExpectExec(deleteSwiftCodeQuery).
			WillReturnError(errors.New("invalid swift code"))
		err = DeleteSwiftCode(db, "")
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db returns error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer db.Close()
		mock.ExpectExec(deleteSwiftCodeQuery).
			WillReturnError(errors.New("connection lost"))
		err = DeleteSwiftCode(db, "PKOPPLPWXXX")
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

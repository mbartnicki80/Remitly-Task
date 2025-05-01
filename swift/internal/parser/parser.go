package parser

import (
	"github.com/xuri/excelize/v2"
	"strings"
)

type SwiftRecord struct {
	SwiftCode     string
	ISO2Code      string
	BankName      string
	Address       string
	Country       string
	IsHeadquarter bool
}

func ParseFromExcel(path string) ([]SwiftRecord, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	var result []SwiftRecord
	for i, record := range rows {
		if i == 0 {
			continue
		}
		isHeadquarter := strings.HasSuffix(record[1], "XXX")

		result = append(result, SwiftRecord{
			ISO2Code:      record[0],
			SwiftCode:     record[1],
			BankName:      record[3],
			Address:       record[4],
			Country:       record[6],
			IsHeadquarter: isHeadquarter,
		})
	}

	return result, nil
}

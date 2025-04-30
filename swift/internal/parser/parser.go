package parser

import (
	"github.com/xuri/excelize/v2"
)

type SwiftRecord struct {
	Code     string
	ISO2     string
	Type     string
	Name     string
	Address  string
	Town     string
	Country  string
	TimeZone string
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
		result = append(result, SwiftRecord{
			ISO2:     record[0],
			Code:     record[1],
			Type:     record[2],
			Name:     record[3],
			Address:  record[4],
			Town:     record[5],
			Country:  record[6],
			TimeZone: record[7],
		})
	}

	return result, nil
}

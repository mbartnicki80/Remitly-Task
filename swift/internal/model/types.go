package model

type SwiftCode struct {
	Address       string `json:"address"`
	BankName      string `json:"bankName"`
	CountryISO2   string `json:"countryISO2"`
	CountryName   string `json:"countryName"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	SwiftCode     string `json:"swiftCode"`
}

type SwiftCodeWithBranches struct {
	SwiftCode
	Branches []SwiftCode `json:"branches,omitempty"`
}

type CountrySwiftCodes struct {
	CountryISO2 string      `json:"countryISO2"`
	CountryName string      `json:"countryName"`
	SwiftCodes  []SwiftCode `json:"swiftCodes"`
}

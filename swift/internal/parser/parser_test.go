package parser

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseFromExcel(t *testing.T) {
	records, err := ParseFromExcel("../../swift_codes.xlsx")
	require.NoError(t, err)
	require.Len(t, records, 1061)

	require.Equal(t, "AL", records[0].ISO2Code)
	require.Equal(t, "AAISALTRXXX", records[0].SwiftCode)
	require.Equal(t, "UNITED BANK OF ALBANIA SH.A", records[0].BankName)
	require.Equal(t, "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023", records[0].Address)
	require.Equal(t, "ALBANIA", records[0].Country)

	require.Equal(t, "AL", records[1060].ISO2Code)
	require.Equal(t, "PYALALT2XXX", records[1060].SwiftCode)
	require.Equal(t, "PAYSERA ALBANIA", records[1060].BankName)
	require.Equal(t, "PALLATI DONIKA, FLOOR KATI 3 RR. FADIL RADA TIRANA, TIRANA, 1001", records[1060].Address)
	require.Equal(t, "ALBANIA", records[1060].Country)
}

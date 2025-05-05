package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mbartnicki80/swift/internal/model"
	"github.com/mbartnicki80/swift/internal/store"
)

func GetSwiftCodeHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		swiftCode := c.Param("swiftCode")
		code, branches, err := store.FetchSwiftCode(db, swiftCode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "SWIFT code not found"})
			return
		}

		if code.IsHeadquarter {
			c.JSON(http.StatusOK, gin.H{
				"address":       code.Address,
				"bankName":      code.BankName,
				"countryISO2":   code.CountryISO2,
				"countryName":   code.CountryName,
				"isHeadquarter": code.IsHeadquarter,
				"swiftCode":     code.SwiftCode,
				"branches":      branches,
			})
		} else {
			c.JSON(http.StatusOK, code)
		}
	}
}

func GetSwiftCodesByCountryHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		iso2 := c.Param("countryISO2")
		codes, err := store.FetchSwiftCodesByCountry(db, iso2)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if len(codes) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No SWIFT codes found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"countryISO2": iso2,
			"countryName": codes[0].CountryName,
			"swiftCodes":  codes,
		})
	}
}

func CreateSwiftCodeHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var record model.SwiftCode
		if err := c.ShouldBindJSON(&record); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		err := store.InsertNewSwiftCode(db, record)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert SWIFT code"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "SWIFT code inserted successfully"})
	}
}

func DeleteSwiftCodeHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		swiftCode := c.Param("swiftCode")
		err := store.DeleteSwiftCode(db, swiftCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete SWIFT code"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "SWIFT code deleted successfully"})
	}
}

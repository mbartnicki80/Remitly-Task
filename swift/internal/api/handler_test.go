package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/mbartnicki80/swift/internal/model"
	"github.com/mbartnicki80/swift/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func clearTables(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("DELETE FROM branches")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM swift_codes")
	require.NoError(t, err)
}

func setupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/v1")
	{
		swift := v1.Group("/swift-codes")
		{
			swift.GET("/:swiftCode", GetSwiftCodeHandler(db))
			swift.GET("/country/:countryISO2", GetSwiftCodesByCountryHandler(db))
			swift.POST("/", CreateSwiftCodeHandler(db))
			swift.DELETE("/:swiftCode", DeleteSwiftCodeHandler(db))
		}
	}
	return r
}

func setupTestDB(t *testing.T) *sql.DB {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("error loading .env")
	}

	dbHost := os.Getenv("TEST_DB_HOST")
	dbPort := os.Getenv("TEST_DB_PORT")
	dbUser := os.Getenv("TEST_DB_USER")
	dbPass := os.Getenv("TEST_DB_PASSWORD")
	dbName := os.Getenv("TEST_DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("cannot connect to test DB: %v", err)
	}

	return db
}

func TestCreateAndGetAndDeleteSwiftCode(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := setupRouter(db)

	payload := `{
		"swiftCode": "TESTCODE123",
		"address": "TEST",
		"countryName": "TEST",
		"countryISO2": "TS",
		"isHeadquarter": true,
		"bankName": "TEST"
	}`
	req := httptest.NewRequest("POST", "/v1/swift-codes/", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	t.Run("Get Swift Code - Found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/v1/swift-codes/TESTCODE123", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "TESTCODE123", response["swiftCode"])
		assert.Equal(t, "TS", response["countryISO2"])
		assert.Equal(t, "TEST", response["countryName"])
	})

	t.Run("Get Swift Code - Not Found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/v1/swift-codes/INVALIDCODE", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "SWIFT code not found", response["error"])
	})

	req = httptest.NewRequest("DELETE", "/v1/swift-codes/TESTCODE123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for DELETE, got %d", w.Code)
	}
}

func TestGetSwiftCodesByCountry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := setupRouter(db)

	deCode1 := model.SwiftCode{SwiftCode: "DEUTDEFFXXX", CountryISO2: "DE", BankName: "Deutsche Bank HQ", Address: "Frankfurt", CountryName: "Germany", IsHeadquarter: true}
	deCode2 := model.SwiftCode{SwiftCode: "DEUTDEFF500", CountryISO2: "DE", BankName: "Deutsche Bank Branch", Address: "Berlin", CountryName: "Germany", IsHeadquarter: false}
	frCode1 := model.SwiftCode{SwiftCode: "BNPAFRPPXXX", CountryISO2: "FR", BankName: "BNP Paribas", Address: "Paris", CountryName: "France", IsHeadquarter: true}

	t.Run("Get Swift Codes By CountryISO2 - Found multiple", func(t *testing.T) {
		clearTables(t, db)
		require.NoError(t, store.InsertNewSwiftCode(db, deCode1))
		require.NoError(t, store.InsertNewSwiftCode(db, deCode2))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/v1/swift-codes/country/DE", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, deCode1.CountryISO2, response["countryISO2"])
		assert.Equal(t, deCode1.CountryName, response["countryName"])

		codes, ok := response["swiftCodes"].([]interface{})
		assert.True(t, ok, "swiftCodes field should be an array")
		assert.Len(t, codes, 2, "Should find 2 codes for DE")
	})

	t.Run("Get Swift Codes By CountryISO2 - Found one", func(t *testing.T) {
		clearTables(t, db)
		require.NoError(t, store.InsertNewSwiftCode(db, frCode1))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/v1/swift-codes/country/FR", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, frCode1.CountryISO2, response["countryISO2"])
		assert.Equal(t, frCode1.CountryName, response["countryName"])

		codes, ok := response["swiftCodes"].([]interface{})
		assert.True(t, ok, "swiftCodes field should be an array")
		assert.Len(t, codes, 1, "Should find 1 code for FR")
	})

	t.Run("Swift Codes By Country - Not Found", func(t *testing.T) {
		clearTables(t, db)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/v1/swift-codes/country/XY", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "No SWIFT codes found", response["error"])
	})
}

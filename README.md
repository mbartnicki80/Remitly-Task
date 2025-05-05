Instructions to run:
1. Create .env file in /swift dir 
   (DB_HOST must be db!) <br />
Example: <br />
DB_HOST=db <br />
DB_PORT=5432 <br />
DB_USER=user <br />
DB_PASSWORD=pass123 <br />
DB_NAME=swift <br />
TEST_DB_HOST=localhost <br />
TEST_DB_PORT=5433 <br />
TEST_DB_USER=user <br />
TEST_DB_PASSWORD=testpass123 <br />
TEST_DB_NAME=testswift <br />

2. Run command: docker-compose up --build <br />

Application should be accessible locally now at port 8080.
You can use HTTP methods such as GET, POST, DELETE using tools like curl. <br />
Examples: <br />
curl -X GET http://localhost:8080/v1/swift-codes/AAISALTRXXX <br />
curl -X POST http://localhost:8080/v1/swift-codes -H "Content-Type: application/json" -d "{\"swiftCode\":\"DEUTDEFFXXX\",\"address\":\"Neue Mainzer Straße 32-36\",\"countryName\":\"Germany\",\"countryISO2\":\"DE\",\"isHeadquarter\":true,\"bankName\":\"Deutsche Bank\"}" <br />
curl -X DELETE http://localhost:8080/v1/swift-codes/NEWCODE123 <br />
curl -X GET http://localhost:8080/v1/swift-codes/country/CL


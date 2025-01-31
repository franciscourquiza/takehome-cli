package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"github.com/olekukonko/tablewriter"
	_ "github.com/denisenkom/go-mssqldb"
)

// We define and adapt the struct for properties based on the challenge requirements
type Property struct {
	SquareFootage int
	Lighting      string
	Price         float64
	Rooms         int
	Bathrooms     int
	Latitude      float64
	Longitude     float64
	Description   string
	Yard          bool
	Garage        bool
	Pool          bool
}

var db *sql.DB //We create the neccesary vars for the string connection to the SQL Server database deployed on Azure
var server = "propertiesdb.database.windows.net"
var port = 1433
var user = "notjohndoe"
var password = "johndoe1$"
var database = "PropertiesDB"

func main() {
	//Creates the connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		server, user, password, port, database)
	var err error
	//Creates the connection pool
	db, err = sql.Open("sqlserver", connString)
	//In case the connection to the database fails
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}

	ctx := context.Background() // => it ensures theres no additional cancellation or timeout applied to the ping
	//Verifies the connection by sending a ping request to the DB
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	//Print the message if the connection was established correctly
	fmt.Printf("Connected!")

	for {
		fmt.Println("Welcome to the property CLI!")
		//Build a query based on user input
		query, args := buildQuery()

		//Executes the query
		rows, err := db.Query(query, args...)
		if err != nil {
			log.Fatalf("Query execution failed: %v", err)
		}
		defer rows.Close()

		//Fetch and display the results
		displayResults(rows)

		// Ask the user if they want to perform another search
		fmt.Print("Do you want to perform another search? (Y/N): ")
		var response string
		fmt.Scanln(&response) 

		if response != "N" && response != "n" && response != "Y" && response != "y" {
			fmt.Println("You have to type Y or N (lowercase is accepted).")
			fmt.Scanln(&response)
		}

		if response == "N" || response == "n" {
			fmt.Println("Goodbye!")
			break
		}
	}
}

//This function constructs the SQL query based on user input
func buildQuery() (string, []interface{}) {
	var filters []string
	var args []interface{}

	prompts := []struct {
		Label string
		Key string
	}{
		{"Square Footage (e.g., > 1000)", "SquareFootage"},
		{"Lighting (low|medium|high)", "Lighting"},
		{"Price (e.g., < 500000)", "Price"},
		{"Rooms (e.g., >= 3)", "Rooms"},
		{"Bathrooms (e.g., = 2)", "Bathrooms"},
		{"Latitude (e.g., = 2)", "Latitude"},
		{"Longitude (e.g., = 2)", "Longitude"},
		{"Description (e.g., = 2)", "Description"},
		{"Yard (If Yes( = 'true') / If No( = 'false'))", "Yard"},
		{"Garage (If Yes( = 'true') / If No( = 'false'))", "Garage"},
		{"Pool (If Yes( = 'true') / If No( = 'false'))", "Pool"},
	}

	for _, prompt := range prompts {
		fmt.Printf("Filter by %s (or leave blank): ", prompt.Label)
		var input string
		fmt.Scanln(&input) //Scan input from the user

		if input != "" {
			filters = append(filters, fmt.Sprintf("%s %s", prompt.Key, input))
			args = append(args, input)
		}
	}

	//Combine filters into WHERE clause
	query := "SELECT SquareFootage, Lighting, Price, Rooms, Bathrooms, Latitude, Longitude, Description, Yard, Garage, Pool FROM Property"
	if len(filters) > 0 {
		query += " WHERE " + filters[0]
		for i := 1; i < len(filters); i++ {
			query += " AND " + filters[i]
		}
	}

	return query, args
}

func displayResults(rows *sql.Rows) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"SquareFt", "Lighting", "Price", "Rooms", "Bathrooms", "Lat", "Lng", "Description", "Yard", "Garage", "Pool"})

	for rows.Next() {
		var prop Property
		if err := rows.Scan(&prop.SquareFootage, &prop.Lighting, &prop.Price, &prop.Rooms, &prop.Bathrooms,
			&prop.Latitude, &prop.Longitude, &prop.Description, &prop.Yard, &prop.Garage, &prop.Pool); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		//We convert the boolean values for Yard, Garage and Pool to Yes or No
		yard := "No"
		if prop.Yard {
			yard = "Yes"
		}
		garage := "No"
		if prop.Garage {
			garage = "Yes"
		}
		pool := "No"
		if prop.Pool {
			pool = "Yes"
		}

		table.Append([]string{
			strconv.Itoa(prop.SquareFootage),
			prop.Lighting,
			fmt.Sprintf("%.2f", prop.Price),
			strconv.Itoa(prop.Rooms),
			strconv.Itoa(prop.Bathrooms),
			fmt.Sprintf("%.4f", prop.Latitude),
			fmt.Sprintf("%.4f", prop.Longitude),
			prop.Description,
			yard,
			garage,
			pool,
		})
	}
	table.Render()
}
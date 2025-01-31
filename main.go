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

	ctx := context.Background() //Creates the context required for the ping
	//Verifies the connection by sending a ping request to the DB, if there's no existing connection it will establish one
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
		defer rows.Close() //Once the query results are no longer needed when an error appears, rows.close must be called to release database resources

		//Fetch and display the results
		displayResults(rows)

		//Ask the user if they want to perform another search
		fmt.Print("Do you want to perform another search? (Y/N): ")
		var response string
		fmt.Scanln(&response) 

		//Shows a message if the user input is different from Y or N
		if response != "N" && response != "n" && response != "Y" && response != "y" {
			fmt.Println("You have to type Y or N (lowercase is accepted).")
			fmt.Scanln(&response)
		}
		//Say goodbye if the user doesn't want to keep searching properties
		if response == "N" || response == "n" {
			fmt.Println("Goodbye!")
			break
		}
	}
}

//This function constructs the SQL query based on user input
func buildQuery() (string, []interface{}) {   //This function returns a string(that is the sql query) and a slice of interface(to hold the values that correspond to our query string placeholders)
	var filters []string                      //Here we create the string that is going to be the sql query
	var args []interface{}				      //And here we create the slice of interface for the placeholders

	prompts := []struct {                     //Here we define a slice of structs, each struct represents a prompt with two fields
		Label string						  //Label describes what kind of input is expected
		Key string							  //Key will match the corresponding database field
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

	for _, prompt := range prompts {             //For each prompt display a message asking the user filter
		fmt.Printf("Filter by %s (or leave blank): ", prompt.Label)
		var input string
		fmt.Scanln(&input) //Scan input from the user

		if input != "" {               
			filters = append(filters, fmt.Sprintf("%s %s", prompt.Key, input))  //If the user applies some filter, append it to the filters slice using the Key and the user Input(for example, SquareFootage > 1000)
			args = append(args, input)       //Stores the raw input into the args slice that we are going to need for the query placeholders
		}
	}

	//Combine filters into WHERE clause
	query := "SELECT SquareFootage, Lighting, Price, Rooms, Bathrooms, Latitude, Longitude, Description, Yard, Garage, Pool FROM Property"
	if len(filters) > 0 {
		query += " WHERE " + filters[0]   //First we require to type WHERE so the query works correctly
		for i := 1; i < len(filters); i++ {
			query += " AND " + filters[i]  //After the WHERE, we start to add the AND word and each filter
		}
	}

	return query, args
}

func displayResults(rows *sql.Rows) {        //This function constructs the table that shows the query results in the console
	table := tablewriter.NewWriter(os.Stdout) //Creates a new tablewriter.Writer instance(from TableWriter library) and it will be printed to the standard output, the terminal
	table.SetHeader([]string{"SquareFt", "Lighting", "Price", "Rooms", "Bathrooms", "Lat", "Lng", "Description", "Yard", "Garage", "Pool"}) //Table header

	for rows.Next() {    //rows.Next is going to verify if there is another row to scan after the first one, in false case, the for loop will finish
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
		//Here we append a slice of strings to the table, each string will represent a field of a database row
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
package main

import (
	_ "github.com/denisenkom/go-mssqldb"
	"context"
	"database/sql"
	"fmt"
	"log"
)

//We define and adapt the struct for properties based on the challenge requirements
type Property struct {
	SquareFootage int
	Lighting string
	Price float64
	Rooms int
	Bathrooms int
	Latitude float64
	Longitude float64
	Description string
	Yard bool
	Garage bool
	Pool bool
}

var db *sql.DB          //We create the neccesary vars for the string connection to the SQL Server database deployed on Azure
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
	
	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	//Print the message if the connection was established correctly
	fmt.Printf("Connected to the database!")
}
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"log"
	"os"
)

// SS_AZ_DSA_DEV:
//     driver: '{ODBC Driver 17 for SQL Server}'
//     server: 'asdbs-de-scus-dsa-dev.database.windows.net'
//     user: 'dsaOwner'
//     password: 'Sw97IwLTriji1r'
//     database: 'db-de-scus-dsa-dev'
//     conn_str: 'mssql+pyodbc:///?odbc_connect={v_params_str}'

func main() {
	// Define flags
	new_connection := flag.Bool("new-connection", false, "Create a new connection")
	nc := flag.Bool("nc", false, "Create a new connection")

	// Parse command-line arguments
	flag.Parse()

	if *new_connection || *nc {
		var db *sql.DB
		var server = "asdbs-de-scus-dsa-dev.database.windows.net"
		var port = 1433
		var user = "dsaOwner"
		var password = "Sw97IwLTriji1r"
		var database = "db-de-scus-dsa-dev"

		// fmt.Println("Please enter server name: ")
		// fmt.Scan(&server)
		// fmt.Println("Please enter user name: ")
		// fmt.Scan(&user)
		// fmt.Println("Please enter password: ")
		// fmt.Scan(&password)
		// Build connection string
		connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
			server, user, password, port, database)
		var err error
		// Create connection pool
		db, err = sql.Open("sqlserver", connString)
		if err != nil {
			fmt.Println("Error creating connection pool: ", err.Error())
		}
		ctx := context.Background()
		err = db.PingContext(ctx)
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Println("Connected!")

		tsql := `
		    SELECT 
			OBJECT_NAME(s.object_id) AS TableName,
			SUM(user_seeks + user_scans + user_lookups) AS TotalAccesses
		    FROM 
			sys.dm_db_index_usage_stats AS s
		    INNER JOIN 
			sys.objects AS o ON s.object_id = o.object_id
		    WHERE 
			OBJECTPROPERTY(s.object_id,'IsUserTable') = 1
		    GROUP BY 
			OBJECT_NAME(s.object_id)
		`

		// Execute query
		rows, err := db.QueryContext(ctx, tsql)
		if err != nil {
			log.Fatal("Error executing query: ", err.Error())
		}
		defer rows.Close()

		// Iterate through the result set
		for rows.Next() {
			var tableName string
			var totalAccesses int

			// Get values from row
			err := rows.Scan(&tableName, &totalAccesses)
			if err != nil {
				log.Fatal("Error scanning row: ", err.Error())
			}

			fmt.Printf("Table Name: %s, Total Accesses: %d\n", tableName, totalAccesses)
		}

	} else {
		fmt.Println("Usage: go run main.go -new-connection")
		os.Exit(1)
	}
}

package main

import (
	"database/sql"
	"fmt"
)

func getUser(db *sql.DB, id string) {
	// VULNERABLE: Direct interpolation via fmt.Sprintf
	query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", id)
	db.Query(query)
}

func main() {
	fmt.Println("Vulnerable Go code")
}

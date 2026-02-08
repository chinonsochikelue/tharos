package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/tls"
	"database/sql"
	"fmt"
	"os/exec"
)

func main() {
	// 1. Command Injection (os/exec)
	userInput := "some; rm -rf /"
	cmd := exec.Command("bash", "-c", "echo "+userInput)
	cmd.Run()

	// 2. Insecure TLS
	_ = &tls.Config{
		InsecureSkipVerify: true,
	}

	// 3. Weak Cryptography
	_ = md5.New()
	_ = sha1.New()

	// 4. SQL Injection (database/sql)
	db, _ := sql.Open("mysql", "user:pass@/dbname")
	id := "123 OR 1=1"
	query := "SELECT * FROM users WHERE id = " + id
	db.Query(query)

	// 5. Hardcoded Secret
	apiKey := "THAROS_SECRET_KEY_9x0y2z_ENTROPY_TEST"
	fmt.Println(apiKey)
}

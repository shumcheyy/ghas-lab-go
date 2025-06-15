package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Hardcoded secret - GHAS secret scanning should catch this
const githubToken = "ghp_exampleinsecuretoken1234567890"

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/exec", execHandler)
	http.HandleFunc("/query", queryHandler)

	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	if username == "admin" && password == "password" {
		fmt.Fprintf(w, "Welcome, admin!")
	} else {
		fmt.Fprintf(w, "Invalid credentials")
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	defer file.Close()

	out, err := os.Create("/tmp/" + header.Filename) // No sanitization - path traversal possible
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	defer out.Close()

	io.Copy(out, file)
	fmt.Fprintf(w, "Uploaded file: %s", header.Filename)
}

func execHandler(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Query().Get("cmd")
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput() // Command Injection possible
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	fmt.Fprintf(w, "Output: %s", out)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("search")
	db, _ := sql.Open("sqlite3", ":memory:")
	db.Exec("CREATE TABLE users (name TEXT)")
	db.Exec("INSERT INTO users (name) VALUES ('alice'), ('bob'), ('admin')")

	// SQL Injection vulnerable
	query := fmt.Sprintf("SELECT name FROM users WHERE name = '%s'", input)
	rows, err := db.Query(query)
	if err != nil {
		fmt.Fprintf(w, "Query error: %v", err)
		return
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}
	fmt.Fprintf(w, "Found: %s", strings.Join(names, ", "))
}

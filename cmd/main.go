package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	simulator "github.com/dmpinheiro/transaction-simulator/internal"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))


	db, err := sql.Open("sqlite3", "file:simulator.db?_journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sim := simulator.NewSimulator(db)

	if err := sim.InitSchema(); err != nil {
		log.Fatal("schema:", err)
	}

	accounts := []string{"A1", "A2", "A3", "A4", "A5", "A6", "A7", "A8", "A9", "A10"}
	if err := sim.SeedAccounts(accounts, 1000); err != nil {
		log.Fatal("seeding:", err)
	}

	// 5 goroutines, each doing 20 transactions
	sim.RunConcurrentTransactions(5, 20)

	sim.PrintAccounts()
	fmt.Println()
	sim.PrintTransactions()
}

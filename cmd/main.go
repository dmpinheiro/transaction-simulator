package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/dmpinheiro/transaction-simulator/config"
	simulator "github.com/dmpinheiro/transaction-simulator/internal"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	config := config.New()


	db, err := sql.Open("sqlite3", "file:simulator.db?_journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sim := simulator.NewSimulator(db)

	if err := sim.InitSchema(); err != nil {
		log.Fatal("schema:", err)
	}

	numAccounts := config.GetInt("num_accounts");
	log.Printf("Seeding %d accounts", numAccounts);
	accounts := make([]string, numAccounts)
	for i := range accounts {
		accounts[i] = fmt.Sprintf("A%d", i+1)
	}
	if err := sim.SeedAccounts(accounts, 1000); err != nil {
		log.Fatal("seeding:", err)
	}

	sim.RunConcurrentTransactions(5, config.GetInt("transactions_per_goroutine"))

	sim.PrintAccounts()
	fmt.Println()
	sim.PrintTransactions()
}

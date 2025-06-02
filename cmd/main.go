package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Account struct {
	ID      string
	Balance int
}

type Transaction struct {
	FromID string
	ToID   string
	Amount int
	Time   time.Time
}

type Simulator struct {
	DB *sql.DB
}

func NewSimulator(db *sql.DB) *Simulator {
	return &Simulator{DB: db}
}

func (s *Simulator) InitSchema() error {
	_, err := s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			balance INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_id TEXT,
			to_id TEXT,
			amount INTEGER,
			time TEXT
		);
	`)
	return err
}

func (s *Simulator) SeedAccounts(ids []string, initialBalance int) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO accounts (id, balance) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.Exec(id, initialBalance); err != nil {
			return err
		}
	}
	return nil
}

func (s *Simulator) GenerateTransaction() error {
	// Get all accounts
	rows, err := s.DB.Query("SELECT id, balance FROM accounts")
	if err != nil {
		return err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Balance); err != nil {
			return err
		}
		accounts = append(accounts, a)
	}
	if len(accounts) < 2 {
		return fmt.Errorf("not enough accounts")
	}

	from := accounts[rand.Intn(len(accounts))]
	to := accounts[rand.Intn(len(accounts))]
	for from.ID == to.ID {
		to = accounts[rand.Intn(len(accounts))]
	}

	if from.Balance <= 0 {
		return fmt.Errorf("insufficient funds in %s", from.ID)
	}

	amount := rand.Intn(from.Balance) + 1
	now := time.Now()

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, from.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, to.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO transactions (from_id, to_id, amount, time) VALUES (?, ?, ?, ?)",
		from.ID, to.ID, amount, now.Format(time.RFC3339),
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *Simulator) PrintAccounts() error {
	rows, err := s.DB.Query("SELECT id, balance FROM accounts")
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Account Balances:")
	for rows.Next() {
		var id string
		var balance int
		rows.Scan(&id, &balance)
		fmt.Printf("  %s: %d\n", id, balance)
	}
	return nil
}

func (s *Simulator) PrintTransactions() error {
	rows, err := s.DB.Query("SELECT from_id, to_id, amount, time FROM transactions ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Transactions:")
	for rows.Next() {
		var fromID, toID, timestamp string
		var amount int
		rows.Scan(&fromID, &toID, &amount, &timestamp)
		fmt.Printf("[%s] %s -> %s: %d\n", timestamp, fromID, toID, amount)
	}
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	db, err := sql.Open("sqlite3", "./simulator.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sim := NewSimulator(db)

	if err := sim.InitSchema(); err != nil {
		log.Fatal("Init schema failed:", err)
	}

	accountIDs := []string{"A1", "A2", "A3", "A4", "A5"}
	if err := sim.SeedAccounts(accountIDs, 1000); err != nil {
		log.Fatal("Seeding accounts failed:", err)
	}

	for i := 0; i < 10; i++ {
		if err := sim.GenerateTransaction(); err != nil {
			log.Println("Transaction skipped:", err)
		}
	}

	if err := sim.PrintAccounts(); err != nil {
		log.Println("Error printing accounts:", err)
	}
	fmt.Println()
	if err := sim.PrintTransactions(); err != nil {
		log.Println("Error printing transactions:", err)
	}
}

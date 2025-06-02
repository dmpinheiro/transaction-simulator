package simulator

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)


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
	for _, id := range ids {
		_, err := s.DB.Exec(`INSERT OR IGNORE INTO accounts (id, balance) VALUES (?, ?)`, id, initialBalance)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Simulator) GenerateTransaction() error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id, balance FROM accounts")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	var accounts []struct {
		ID      string
		Balance int
	}
	for rows.Next() {
		var id string
		var balance int
		if err := rows.Scan(&id, &balance); err != nil {
			tx.Rollback()
			return err
		}
		accounts = append(accounts, struct {
			ID      string
			Balance int
		}{id, balance})
	}

	if len(accounts) < 2 {
		tx.Rollback()
		return fmt.Errorf("not enough accounts")
	}

	from := accounts[rand.Intn(len(accounts))]
	to := accounts[rand.Intn(len(accounts))]
	for from.ID == to.ID {
		to = accounts[rand.Intn(len(accounts))]
	}

	if from.Balance <= 0 {
		tx.Rollback()
		return fmt.Errorf("insufficient funds in %s", from.ID)
	}

	amount := rand.Intn(from.Balance) + 1
	now := time.Now().Format(time.RFC3339)

	_, err = tx.Exec(`UPDATE accounts SET balance = balance - ? WHERE id = ?`, amount, from.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(`UPDATE accounts SET balance = balance + ? WHERE id = ?`, amount, to.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(`INSERT INTO transactions (from_id, to_id, amount, time) VALUES (?, ?, ?, ?)`,
		from.ID, to.ID, amount, now)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *Simulator) RunConcurrentTransactions(workers, perWorker int) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				// Retry logic for SQLite write locks
				for retries := 0; retries < 3; retries++ {
					err := s.GenerateTransaction()
					if err == nil {
						break
					}
					if retries == 2 {
						log.Printf("[worker %d] failed transaction after 3 retries: %v", workerID, err)
					}
					time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
				}
			}
		}(i)
	}
	wg.Wait()
}

func (s *Simulator) PrintAccounts() {
	rows, _ := s.DB.Query("SELECT id, balance FROM accounts")
	defer rows.Close()
	fmt.Println("Account Balances:")
	for rows.Next() {
		var id string
		var balance int
		rows.Scan(&id, &balance)
		fmt.Printf("  %s: %d\n", id, balance)
	}
}

func (s *Simulator) PrintTransactions() {
	rows, _ := s.DB.Query("SELECT from_id, to_id, amount, time FROM transactions ORDER BY id")
	defer rows.Close()
	fmt.Println("Transactions:")
	for rows.Next() {
		var fromID, toID, ts string
		var amount int
		rows.Scan(&fromID, &toID, &amount, &ts)
		fmt.Printf("[%s] %s -> %s: %d\n", ts, fromID, toID, amount)
	}
}

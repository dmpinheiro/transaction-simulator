package simulator

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/dmpinheiro/transaction-simulator/internal/domain"
	"gorm.io/gorm"
)

type Simulator struct {
	DB *gorm.DB
}

func NewSimulator(db *gorm.DB) *Simulator {
	return &Simulator{DB: db}
}

func (s *Simulator) InitSchema() error {
	s.DB.AutoMigrate(&domain.Account{}, &domain.Transaction{})
	return nil
}

func (s *Simulator) SeedAccounts(ids []string, initialBalance int) error {

	for _, id := range ids {
		result := s.DB.Create(&domain.Account{ID: id, Balance: initialBalance})
		err := result.Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Simulator) GenerateTransaction() error {

	tx := s.DB.Begin()
	var accounts []domain.Account
	result := s.DB.Find(&accounts)
	if result.Error != nil {
		log.Printf("Error fetching accounts: %v", result.Error)
		return result.Error
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

	result = tx.Raw(`UPDATE accounts SET balance = balance - ? WHERE id = ?`, amount, from.ID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}
	result = tx.Raw(`UPDATE accounts SET balance = balance + ? WHERE id = ?`, amount, to.ID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}
	result = tx.Raw(`INSERT INTO transactions (from_id, to_id, amount, time) VALUES (?, ?, ?, ?)`,
		from.ID, to.ID, amount, now)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	result = tx.Commit()
	if result.Error != nil {
		log.Printf("Error committing transaction: %v", result.Error)
		return result.Error
	}
	return nil
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
	var accounts []domain.Account
	result := s.DB.Find(&accounts)
	if result.Error != nil {
		log.Printf("Error fetching accounts: %v", result.Error)
		return
	}

	fmt.Println("Account Balances:")

	for _, account := range accounts {
		fmt.Printf("  %s: %d\n", account.ID, account.Balance)
	}
}

func (s *Simulator) PrintTransactions() {
	var transactions []domain.Transaction
	s.DB.Raw("SELECT from_id, to_id, amount, time FROM transactions ORDER BY id").Scan(&transactions)
	fmt.Println("Transactions:")
	for _, transaction := range transactions {
		fmt.Printf("[%s] %s -> %s: %d\n", transaction.AccountID, transaction.FromID, transaction.ToID, transaction.Amount)
	}
}

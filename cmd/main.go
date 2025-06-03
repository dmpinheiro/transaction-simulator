package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmpinheiro/transaction-simulator/config"
	simulator "github.com/dmpinheiro/transaction-simulator/internal"
	"github.com/dmpinheiro/transaction-simulator/internal/infrastructure"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	config := config.New()

	app := infrastructure.NewFiber(config)
	port := config.GetInt("app.port")

	db := infrastructure.NewGorm(config)

	sim := simulator.NewSimulator(db)
	sim.InitSchema()

	numAccounts := config.GetInt("num_accounts")
	log.Printf("Seeding %d accounts", numAccounts)
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

	go func() {
		if err := app.Listen(fmt.Sprintf(":%v", port)); err != nil {
			panic(fmt.Errorf("error running app : %+v", err.Error()))
		}
	}()

	ch := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel

	<-ch // This blocks the main thread until an interrupt is received

	// Your cleanup tasks go here
	_ = app.Shutdown()

	fmt.Println("App was successful shutdown.")

}

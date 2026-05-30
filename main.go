package main

import (
	"fmt"
	"log"
	"os"

	"github.com/networkchaos/BRD_Currency/internal/detector"
	"github.com/networkchaos/BRD_Currency/internal/fetcher"
	"github.com/networkchaos/BRD_Currency/internal/notifier"
	"github.com/networkchaos/BRD_Currency/internal/state"
)

func main() {
	fmt.Println("Starting BRD Currency Detector...")

	//these come from environment variables (Github Actions secrets later
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

	if telegramToken == "" || telegramChatID == "" {
		log.Fatal("TELEGRAM_TOKEN and TELEGRAM_CHAT_ID environment variables must be set")
	}

	//currency we want to watch (base : USD )
	watchList := []string{"IRR", "ZWL", "VES", "TRY", "NGN"}

	//step 1  : fetch current rates
	fmt.Println("Fetching current exchange rates...")
	rates, err := fetcher.GetRates(watchList)
	if err != nil {
		log.Fatalf("failed to fetch rate : %v", err)

	}
	fmt.Printf("Got rate for %d currencies\n", len(rates))

	//step2 : load previous rates from state file
	previous, err := state.Load("state.json")
	if err != nil {
		fmt.Println("no previous state found , saving current as baseline..")
		if saveErr := state.Save("state.json", rates); saveErr != nil {
			log.Fatalf("failed to save state : %v", saveErr)
		}
		fmt.Println("Baseline saved run again tommorow alerts")
		return
	}
	//step 3 : detect signals (buy/sell)
	fmt.Println("analysing price changes...")
	signals := detector.Analyse(previous, rates)

	//send notifications
	n := notifier.NewTelegram(telegramToken, telegramChatID)
	for _, sig := range signals {
		fmt.Println("Signal %s\n", sig.Message)
		if err := n.Send(sig.Message); err != nil {
			log.Printf("failed to send notification : %v", err)
		}
	}

	if len(signals) == 0 {
		fmt.Println("No significant changes detected.")
	}

	//step 5 :
	if errr := state.Save("state.json", rates); errr != nil {
		log.Fatalf("failed to save state : %v", errr)
	}
	fmt.Println("State updated successfully.")

}

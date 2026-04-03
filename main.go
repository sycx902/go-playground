package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Transaction mirrors a real-world budget entry.
type Transaction struct {
	Amount      float64 `json:"Amount"`
	Category    string  `json:"Category"`
	Description string  `json:"Description"`
	Date        string  `json:"Date"`
}

func main() {
	if len(os.Args) == 1 {
		interactiveAdd()
		return
	}
	// FORCE MIGRATE DEBUG - Remove after success

	// if err := migrateCSVtoJSON("budget.csv", "budget.json"); err != nil {
		// log.Fatal(err)
	// }



	switch os.Args[1] {
	case "add":
		addTransaction()
	case "list":
		printSummary("budget.json")
	case "delete":
		deleteTransaction()
	case "edit":
		editTransaction()
	default:
		fmt.Println("Usage: go run main.go [add|list|delete|edit]")
		fmt.Println("  add: Interactive or flags (-am=50 -cat=Food etc.)")
		os.Exit(1)
	}
}

func loadTransactions(filename string) ([]Transaction, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []Transaction{}, nil
		}
		return nil, err
	}
	var txs []Transaction
	err = json.Unmarshal(data, &txs)
	return txs, err
}

func saveTransactions(filename string, txs []Transaction) error {
	data, err := json.MarshalIndent(txs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func migrateCSVtoJSON(csvFile, jsonFile string) error {
	f, err := os.Open(csvFile)
	if err != nil {
		fmt.Printf("CSV migration skipped (no %s): %v\n", csvFile, err)
		return nil
	}
	defer f.Close()
	
	fmt.Printf("Migrating %s...\n", csvFile)
	
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("CSV read error: %v\n", err)
		return nil
	}
	
	fmt.Printf("Found %d data rows\n", len(records)-1)
	
	var txs []Transaction
	for i, row := range records[1:] {  // Skip header
		if len(row) < 4 {
			fmt.Printf("Skipping invalid row %d: %v\n", i+1, row)
			continue
		}
		amt, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			fmt.Printf("Skipping row %d (bad amount): %v\n", i+1, row[0])
			continue
		}
		t := Transaction{
			Amount:      amt,
			Category:    row[1],
			Description: row[2],
			Date:        row[3],
		}
		txs = append(txs, t)
	}
	
	if len(txs) == 0 {
		fmt.Println("No valid transactions to migrate")
		return nil
	}
	
	if err := saveTransactions(jsonFile, txs); err != nil {
		return fmt.Errorf("save JSON: %w", err)
	}
	fmt.Printf("Migrated %d transactions to %s\n", len(txs), jsonFile)
	return nil
}

func addTransaction() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addAmount := addCmd.Float64("am", 0, "Amount (±)")
	addCategory := addCmd.String("cat", "", "Category")
	addDesc := addCmd.String("dsc", "", "Description")
	addDate := addCmd.String("dat", "", "Date (YYYY-MM-DD)")
	addCmd.Parse(os.Args[2:])

	t := promptTransaction(*addAmount, *addCategory, *addDesc, *addDate)
	if t.Category == "" {
		log.Fatal("Category required")
	}

	txs, err := loadTransactions("budget.json")
	if err != nil {
		log.Fatal(err)
	}
	migrateCSVtoJSON("budget.csv", "budget.json") // Migrate if exists

	txs = append(txs, t)
	if err := saveTransactions("budget.json", txs); err != nil {
		log.Fatal(err)
	}
	printSummary("budget.json")
}

func interactiveAdd() {
	reader := bufio.NewReader(os.Stdin)
	txs := make([]Transaction, 0)

	for {
		t := promptInteractive(reader)
		if t.Category == "" {
			fmt.Println("Category required. Skipping.")
			continue
		}
		txs = append(txs, t)

		fmt.Print("Add more? (y/n): ")
		more, _ := reader.ReadString('\n')
		trimmed := strings.TrimSpace(more)
		if len(trimmed) > 0 && (trimmed[0] == 'n' || trimmed[0] == 'N') {
			break
		}
	}

	budgetFile := "budget.json"
	existing, err := loadTransactions(budgetFile)
	if err != nil {
		log.Fatal(err)
	}
	migrateCSVtoJSON("budget.csv", budgetFile)

	existing = append(existing, txs...)
	if err := saveTransactions(budgetFile, existing); err != nil {
		log.Fatal(err)
	}
	printSummary(budgetFile)
}

func promptInteractive(reader *bufio.Reader) Transaction {
	t := Transaction{Date: time.Now().Format("2006-01-02")}

	fmt.Print("Amount (±): ")
	amtStr, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	t.Amount, _ = strconv.ParseFloat(strings.TrimSpace(amtStr), 64)

	fmt.Print("Category: ")
	cat, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	t.Category = strings.TrimSpace(cat)

	fmt.Print("Description: ")
	desc, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	t.Description = strings.TrimSpace(desc)

	fmt.Print("Date (YYYY-MM-DD, Enter=today): ")
	dateStr, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	dateTrim := strings.TrimSpace(dateStr)
	if dateTrim != "" {
		t.Date = dateTrim
	}
	return t
}

func promptTransaction(am float64, cat, desc, dat string) Transaction {
	t := Transaction{
		Amount:      am,
		Category:    cat,
		Description: desc,
		Date:        time.Now().Format("2006-01-02"),
	}
	if dat != "" {
		t.Date = dat
	}
	return t
}

func printTransactions(txs []Transaction) {
	if len(txs) == 0 {
		fmt.Println("No transactions.")
		return
	}
	for i, t := range txs {
		sign := ""
		if t.Amount > 0 {
			sign = "+"
		}
		fmt.Printf("%d. %s%.2f | %s | %s | %s\n", 
			i+1, sign, t.Amount, t.Category, t.Description, t.Date)
	}
}

func printSummary(filename string) {
	txs, err := loadTransactions(filename)
	if err != nil || len(txs) == 0 {
		fmt.Println("No transactions yet.")
		return
	}

	var totalIncome, totalExpenses float64
	for _, t := range txs {
		if t.Amount > 0 {
			totalIncome += t.Amount
		} else {
			totalExpenses += t.Amount
		}
	}
	balance := totalIncome + totalExpenses

	fmt.Printf("Transactions (%d):\n", len(txs))
	printTransactions(txs)
	fmt.Printf("\nTotal Income: %.2f\nTotal Expenses: %.2f\nBalance: %.2f\n", totalIncome, totalExpenses, balance)
}

func deleteTransaction() {
	txs, err := loadTransactions("budget.json")
	if err != nil {
		log.Fatal(err)
	}
	migrateCSVtoJSON("budget.csv", "budget.json")

	if len(txs) == 0 {
		fmt.Println("No transactions.")
		return
	}

	printTransactions(txs)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter index to delete: ")
	idxStr, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	idx, err := strconv.Atoi(strings.TrimSpace(idxStr))
	if err != nil || idx < 1 || idx > len(txs) {
		log.Fatal("Invalid index")
	}
	txs = append(txs[:idx-1], txs[idx:]...)
	if err := saveTransactions("budget.json", txs); err != nil {
		log.Fatal(err)
	}
	printSummary("budget.json")
}

func editTransaction() {
	txs, err := loadTransactions("budget.json")
	if err != nil {
		log.Fatal(err)
	}
	migrateCSVtoJSON("budget.csv", "budget.json")

	if len(txs) == 0 {
		fmt.Println("No transactions.")
		return
	}

	printTransactions(txs)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter index to edit: ")
	idxStr, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	idx, err := strconv.Atoi(strings.TrimSpace(idxStr))
	if err != nil || idx < 1 || idx > len(txs) {
		log.Fatal("Invalid index")
	}

	// Edit interactively
	t := promptInteractive(reader)
	if t.Category != "" { // Only update if provided
		txs[idx-1] = t
	} else {
		fmt.Println("No changes (category required for update).")
		return
	}

	if err := saveTransactions("budget.json", txs); err != nil {
		log.Fatal(err)
	}
	printSummary("budget.json")
}

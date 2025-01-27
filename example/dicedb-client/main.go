package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dicedb/dicedb-go"
)

func main() {
	client := dicedb.NewClient(&dicedb.Options{
		Addr:         "localhost:7379",
		Password:     "",
		MinIdleConns: 1,
	})

	ctx := context.Background()

	start := time.Now()

	_, err := client.Ping(ctx).Result()

	fmt.Printf("Ping command is taking %.2f seconds to execute\n", time.Since(start).Seconds())

	if err != nil {
		log.Fatalf("Could not connect to DiceDB: %v", err)
	}
}

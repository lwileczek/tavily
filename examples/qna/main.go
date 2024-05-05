package main

import (
	"fmt"
	"os"

	"github.com/lwileczek/tavily"
)

func main() {
	API_KEY := os.Getenv("TAVILY_API_KEY")
	client, err := tavily.NewClient(API_KEY)
	if err != nil {
		fmt.Println("Oops must have forgot my API key!", err)
		return
	}

	answer, err := client.QASearch("Where does Messi play right now?")
	if err != nil {
		fmt.Println("Unable to get an answer for that, sorry", err)
		return
	}

	fmt.Println(answer)
}

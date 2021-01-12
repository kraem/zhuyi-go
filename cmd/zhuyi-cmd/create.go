package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/kraem/zhuyi-go/pkg/payloads"
	"github.com/kraem/zhuyi-go/network"
)

func writeOutput(fileName *string, err error) {
	r := payloads.NewAppendResponse(fileName, err)
	w := os.Stdout
	if err != nil {
		w = os.Stderr
	}
	json.NewEncoder(w).Encode(r)
}

func main() {
	c, err := network.NewConfig()
	if err != nil {
		writeOutput(nil, err)
		os.Exit(1)
	}

	// TODO help msgs
	var title = flag.String("title", "", "help message for flag n")
	var body = flag.String("body", "", "help message for flag n")
	flag.Parse()

	fileName, err := c.CreateNode(*title, *body)

	writeOutput(&fileName, nil)
}

package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/kraem/zhuyi-go/pkg/payloads"
	"github.com/kraem/zhuyi-go/zettel"
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
	c, err := zettel.NewConfig()
	if err != nil {
		writeOutput(nil, err)
		os.Exit(1)
	}

	// TODO help msgs
	var title = flag.String("title", "", "help message for flag n")
	var body = flag.String("body", "", "help message for flag n")
	flag.Parse()

	fileName, err := c.CreateZettel(*title, *body)

	writeOutput(&fileName, nil)
}

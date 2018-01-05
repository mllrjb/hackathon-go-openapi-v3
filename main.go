package main

import (
	"fmt"
	"os"

	"github.schq.secious.com/jason-miller/go-openapi-v3/generated"
)

func main() {
	// generated.IUpdateCaseHandler_VndLogrhythmCaseV1 = generated.ImplUpdateCaseHandler_VndLogrhythmCaseV1{}

	address := "127.0.0.1:9535"
	server := generated.NewServer("127.0.0.1:9535")

	fmt.Printf("Listening on %s\n", address)
	server.ListenAndServe()

	os.Exit(0)
}

package main

import (
	"fmt"
	"os"

	"github.com/mllrjb/hackathon-go-openapi-v3/generated/component"
	"github.com/mllrjb/hackathon-go-openapi-v3/generated/operation"

	"github.com/mllrjb/hackathon-go-openapi-v3/generated"
)

func main() {
	generated.CreateItemsHandler_VndItem = operation.CreateItemsHandler_VndItemFunc(func(params operation.CreateItemsParameters, body component.Item) operation.Responder {
		return operation.StatusCodeResponder(204)
	})
	generated.CreateItemsHandler_VndItems = operation.CreateItemsHandler_VndItemsFunc(func(params operation.CreateItemsParameters, body []component.Item) operation.Responder {
		return operation.JsonResponder(201, "application/vnd.foo.bar+json", "whatever")
	})
	address := "127.0.0.1:9535"
	server := generated.NewServer("127.0.0.1:9535")

	fmt.Printf("Listening on %s\n", address)
	server.ListenAndServe()

	os.Exit(0)
}

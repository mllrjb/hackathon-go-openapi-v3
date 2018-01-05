package generated

import (
	"net/http"
	"time"
)

// type CustomRouter struct {
// 	Negroni *negroni.Negroni
// }

func NewServer(address string) *http.Server {
	// router, err := CreateCustomRouter()
	// if err != nil {
	// 	panic(err)
	// }

	return &http.Server{
		Handler: CreateAPIRouter(),
		// Handler: router,
		Addr: address,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 120 * time.Second,
		ReadTimeout:  120 * time.Second,
	}
}

// func CreateCustomRouter() (cr *CustomRouter, err error) {
// 	api := negroni.New()
// 	api.UseHandler(CreateAPIRouter())

// 	return &CustomRouter{
// 		Negroni: api,
// 	}, nil
// }

// func (n *CustomRouter) ServeHTTP(res http.ResponseWriter, req *http.Request) {
// 	n.Negroni.ServeHTTP(res, req)
// }

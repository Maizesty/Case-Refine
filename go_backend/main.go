package main

import (
	// "bytes"
	"fmt"
	// "net/http"
	// "net/http/httptest"
	// "net/url"

	"github.com/valyala/fasthttp"
	"github.com/buaazp/fasthttprouter"

)

// HandlePredict is the handler for the /predict path.
func HandlePredict(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "Welcome!\n")
}

// handleKeys is a placeholder for the logic that would process the keys.
// In this example, we just return the keys.
// func handleKeys(keys []int) []int {
// 	return keys
// }

func main() {
	// Register the handler for the /predict path
	fasthttpRouter := fasthttprouter.New()
	fasthttpRouter.POST("/predict", HandlePredict)

	// Start the server
	fmt.Println("Server listening on port 5091")
	if err := fasthttp.ListenAndServe(":5091", fasthttpRouter.Handler); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

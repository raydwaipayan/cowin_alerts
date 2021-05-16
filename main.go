package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/raydwaipayan/cowin_alerts/util"
	"github.com/valyala/fasthttp"
)

func main() {
	godotenv.Load()

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			util.ReceiveWebhook(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	fmt.Println("Starting cowin alerts server on port 8080")
	fasthttp.ListenAndServe(":8080", requestHandler)
}

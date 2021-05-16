package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/raydwaipayan/cowin_alerts/util"
	"github.com/valyala/fasthttp"
)

func main() {
	godotenv.Load()

	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				util.SendUpdates()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			util.ReceiveWebhook(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting cowin alerts server on port %s\n", port)
	fasthttp.ListenAndServe(fmt.Sprintf(":%s", port), requestHandler)
}

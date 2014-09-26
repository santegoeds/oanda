package main

import (
	"fmt"

	"github.com/santegoeds/oanda"
)

func main() {
	client, err := oanda.NewSandboxClient()
	if err != nil {
		panic(err)
	}

	// List available instruments
	instruments, err := client.Instruments(nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(instruments)

	// Buy one unit of EUR/USD with a trailing stop of 10 pips.
	tr, err := client.NewTrade(oanda.Buy, 1, "eur_usd", oanda.TrailingStop(10.0))
	if err != nil {
		panic(err)
	}
	fmt.Println(tr)

	// Create and run a price server.
	priceServer, err := client.NewPriceServer("eur_usd")
	if err != nil {
		panic(err)
	}
	priceServer.ConnectAndHandle(func(instrument string, tick oanda.PriceTick) {
		fmt.Println("Received tick:", instrument, tick)
		priceServer.Stop()
	})
}

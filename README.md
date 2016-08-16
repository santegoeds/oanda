# OANDA for Go

[![Build Status](https://travis-ci.org/santegoeds/oanda.svg?branch=master)](https://travis-ci.org/santegoeds/oanda)
[![Coverage Status](https://coveralls.io/repos/santegoeds/oanda/badge.svg?branch=master&service=github)](https://coveralls.io/github/santegoeds/oanda?branch=master)

OANDA for Go is a library to access the [OANDA REST API](http://developer.oanda.com) from open source programming language [Go](http://www.golang.org).

[OANDA](http://www.oanda.com) is a technology company that provides internet based forex trading and currency information. [OANDA](http://www.oanda.com) provide access to their services via a free 
REST API. [OANDA](http://www.oanda.com) is not affiliated and does not endorse or recommend 
OANDA for Go.

Please note that this is an alpha release and the code is likely to contain bugs.

## Usage

```Go
package main

import (
	"flag"
	"fmt"

	"github.com/santegoeds/oanda"
)

var (
	token   = flag.String("token", "", "Oanda authorization token.")
	account = flag.Int64("account", 0, "Oanda account.")
)

func main() {
	flag.Parse()

	if *token == "" {
		panic("An Oanda authorization token is required")
	}

	if *account == 0 {
		panic("An Oanda account is required")
	}

	client, err := oanda.NewFxPracticeClient(*token)
	if err != nil {
		panic(err)
	}

	client.SelectAccount(oanda.Id(*account))

	// List available instruments
	instruments, err := client.Instruments(nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(instruments)

	// Buy one unit of EUR/USD with a trailing stop of 10 pips.
	tradeInfo, err := client.NewTrade(oanda.Buy, 1, "eur_usd", oanda.TrailingStop(10.0))
	if err != nil {
		panic(err)
	}
	fmt.Println(tradeInfo)

	// Create and run a price server.
	priceServer, err := client.NewPriceServer("eur_usd")
	if err != nil {
		panic(err)
	}
	priceServer.ConnectAndHandle(func(instrument string, tick oanda.PriceTick) {
		fmt.Println("Received tick:", instrument, tick)
		priceServer.Stop()
	})

	// Close the previously opened trade.
	tradeCloseInfo, err := client.CloseTrade(tradeInfo.TradeId)
	if err != nil {
		panic(err)
	}
	fmt.Println(tradeCloseInfo)
}
```

## License

Oanda for Go is released under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

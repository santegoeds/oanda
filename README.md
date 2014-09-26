# OANDA for Go

[![Build Status](https://travis-ci.org/santegoeds/oanda.svg?branch=master)](https://travis-ci.org/santegoeds/oanda)

OANDA for Go is a library to access the [OANDA REST API](http://developer.oanda.com) from open source programming language [Go](http://www.golang.org).

[OANDA](http://www.oanda.com) is a technology company that provides internet based forex trading and currency information. [OANDA](http://www.oanda.com) provide access to their services via a free 
REST API. OANDA for Go is not affiliated or endorsed by [OANDA](http://www.oanda.com).

Please note that this is an alpha release and the code is likely to contain bugs.

## Example

```Go
package main

import (
    "fmt"

    "github.com/santegoeds/oanda"
)
    
client, err := oanda.NewSandboxClient()
if err != nil {
    panic(err)
}

// List available instruments
instruments, err := client.Instruments()
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
priceServer.ConnectAndHandle(func(instrument string, tick *oanda.Tick) {
    fmt.Println("Received tick:", instrument, tick)
})
```

## License

Oanda for Go is released under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

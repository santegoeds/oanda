## Functionality

* Fix Data race when accessing chanMap in Prices & Event servers.
  The potential race condition occurs when the servers modify the map in "finish()" when there are
  still ticks/events being processed.
* Refactor the mess that is Context, Transport, Canceling of Requests and limiting idle connections 
  (the latter of which is only necessary when connecting to the stream servers).
* Event Streaming
* Add support for sessions to streaming Api's.
* Add logging.
* Forex Labs
    * Orderbook
    * Calendar.
    * Historical position rates.
    * Spreads.
    * Commitments of Traders.
 * Add support for using the Status Api.

## Testing

* Test getting all transactions.
* Test error conditions.

## Documentation

* Improve/complete documentation.
* Provide usage examples.

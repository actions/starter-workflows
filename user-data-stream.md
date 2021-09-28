<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [General WSS information](#general-wss-information)
- [API Endpoints](#api-endpoints)
  - [Create a listenKey](#create-a-listenkey)
  - [Ping/Keep-alive a listenKey](#pingkeep-alive-a-listenkey)
  - [Close a listenKey](#close-a-listenkey)
- [Web Socket Payloads](#web-socket-payloads)
  - [Account Update](#account-update)
  - [Balance Update](#balance-update)
  - [Order Update](#order-update)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# User Data Streams for Binance (2021-01-01)
# General WSS information
* The base API endpoint is: **https://api.binance.com**
* A User Data Stream `listenKey` is valid for 60 minutes after creation.
* Doing a `PUT` on an active `listenKey` will extend its validity for 60 minutes.
* Doing a `DELETE` on an active `listenKey` will close the stream and invalidate the `listenKey`.
* Doing a `POST` on an account with an active `listenKey` will return the currently active `listenKey` and extend its validity for 60 minutes.
* The base websocket endpoint is: **wss://stream.binance.com:9443**
* User Data Streams are accessed at **/ws/\<listenKey\>** or **/stream?streams=\<listenKey\>**
* A single connection to **stream.binance.com** is only valid for 24 hours; expect to be disconnected at the 24 hour mark

# API Endpoints
## Create a listenKey
```
POST /api/v3/userDataStream
```
Start a new user data stream. The stream will close after 60 minutes unless a keepalive is sent. If the account has an active `listenKey`, that `listenKey` will be returned and its validity will be extended for 60 minutes.

**Weight:**
1

**Parameters:**
NONE

**Response:**
```javascript
{
  "listenKey": "pqia91ma19a5s61cv6a81va65sdf19v8a65a1a5s61cv6a81va65sdf19v8a65a1"
}
```

## Ping/Keep-alive a listenKey
```
PUT /api/v3/userDataStream
```
Keepalive a user data stream to prevent a time out. User data streams will close after 60 minutes. It's recommended to send a ping about every 30 minutes.

**Weight:**
1

**Parameters:**

Name | Type | Mandatory | Description
------------ | ------------ | ------------ | ------------
listenKey | STRING | YES

**Response:**
```javascript
{}
```

## Close a listenKey
```
DELETE /api/v3/userDataStream
```
Close out a user data stream.

**Weight:**
1

**Parameters:**

Name | Type | Mandatory | Description
------------ | ------------ | ------------ | ------------
listenKey | STRING | YES

**Response:**
```javascript
{}
```

# Web Socket Payloads
## Account Update

`outboundAccountPosition` is sent any time an account balance has changed and contains the assets that were possibly changed by the event that generated the balance change.

```javascript
{
  "e": "outboundAccountPosition", //Event type
  "E": 1564034571105,             //Event Time
  "u": 1564034571073,             //Time of last account update
  "B": [                          //Balances Array
    {
      "a": "ETH",                 //Asset
      "f": "10000.000000",        //Free
      "l": "0.000000"             //Locked
    }
  ]
}
```

## Balance Update

Balance Update occurs during the following:
* Deposits or withdrawals from the account
* Transfer of funds between accounts (e.g. Spot to Margin)

**Payload**
```javascript
{
  "e": "balanceUpdate",         //Event Type
  "E": 1573200697110,           //Event Time
  "a": "BTC",                   //Asset
  "d": "100.00000000",          //Balance Delta
  "T": 1573200697068            //Clear Time
}
```

## Order Update
Orders are updated with the `executionReport` event.

Check the [Rest API Documentation](./rest-api.md#enum-definitions) and below for relevant enum definitions.

Average price can be found by doing `Z` divided by `z`.

**Payload:**
```javascript
{
  "e": "executionReport",        // Event type
  "E": 1499405658658,            // Event time
  "s": "ETHBTC",                 // Symbol
  "c": "mUvoqJxFIILMdfAW5iGSOW", // Client order ID
  "S": "BUY",                    // Side
  "o": "LIMIT",                  // Order type
  "f": "GTC",                    // Time in force
  "q": "1.00000000",             // Order quantity
  "p": "0.10264410",             // Order price
  "P": "0.00000000",             // Stop price
  "F": "0.00000000",             // Iceberg quantity
  "g": -1,                       // OrderListId
  "C": null,                     // Original client order ID; This is the ID of the order being canceled
  "x": "NEW",                    // Current execution type
  "X": "NEW",                    // Current order status
  "r": "NONE",                   // Order reject reason; will be an error code.
  "i": 4293153,                  // Order ID
  "l": "0.00000000",             // Last executed quantity
  "z": "0.00000000",             // Cumulative filled quantity
  "L": "0.00000000",             // Last executed price
  "n": "0",                      // Commission amount
  "N": null,                     // Commission asset
  "T": 1499405658657,            // Transaction time
  "t": -1,                       // Trade ID
  "I": 8641984,                  // Ignore
  "w": true,                     // Is the order on the book?
  "m": false,                    // Is this trade the maker side?
  "M": false,                    // Ignore
  "O": 1499405658657,            // Order creation time
  "Z": "0.00000000",             // Cumulative quote asset transacted quantity
  "Y": "0.00000000",              // Last quote asset transacted quantity (i.e. lastPrice * lastQty)
  "Q": "0.00000000"              // Quote Order Qty
}
```

**Execution types:**

* NEW - The order has been accepted into the engine.
* CANCELED - The order has been canceled by the user. 
* REPLACED (currently unused)
* REJECTED - The order has been rejected and was not processed. (This is never pushed into the User Data Stream)
* TRADE - Part of the order or all of the order's quantity has filled.
* EXPIRED - The order was canceled according to the order type's rules (e.g. LIMIT FOK orders with no fill, LIMIT IOC or MARKET orders that partially fill) or by the exchange, (e.g. orders canceled during liquidation, orders canceled during maintenance)

If the order is an OCO, an event will be displayed named `ListStatus` in addition to the `executionReport` event.

**Payload**
```javascript
{
  "e": "listStatus",                //Event Type
  "E": 1564035303637,               //Event Time
  "s": "ETHBTC",                    //Symbol
  "g": 2,                           //OrderListId
  "c": "OCO",                       //Contingency Type
  "l": "EXEC_STARTED",              //List Status Type
  "L": "EXECUTING",                 //List Order Status
  "r": "NONE",                      //List Reject Reason
  "C": "F4QN4G8DlFATFlIUQ0cjdD",    //List Client Order ID
  "T": 1564035303625,               //Transaction Time
  "O": [                            //An array of objects
    {
      "s": "ETHBTC",                //Symbol
      "i": 17,                      // orderId
      "c": "AJYsMjErWJesZvqlJCTUgL" //ClientOrderId
    },
    {
      "s": "ETHBTC",
      "i": 18,
      "c": "bfYPSQdLoqAJeNrOr9adzq"
    }
  ]
}
```

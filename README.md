Simple location history server
==============================

## To run
Either run or build.
The program looks for the listening port, `LHPORT`, in the environment variables.
If not set it defaults to `8080`.
The client communicates with the server via JSON data.
For example:
```
LHPORT=9090 go run server.go
```

## API endpoints
* `GET /`\
Dumps all the orders and each of their associated locations.
```
curl localhost:8080/
```

* `GET /location`\
Dumps all the orders and each of their associated locations.
```
curl localhost:8080/location/
```

* `PUT /location/{order_id}`\
Adds a new order.
```
curl -X POST -H "Content-Type: application/json" -d '{"order_id":"123"}' localhost:8080/location/
```

* `GET /location/{order_id}?max=<N>`\
Gets the location history for the specified order.
The optional query parameter `max` is used to return the most recent `max` location entries for the specified order.

* `DELETE /location/{order_id}`\
Deletes an order.

## To do
* Add extra server side logging
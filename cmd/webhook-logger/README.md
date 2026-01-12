# Webhook Logger

A simple HTTP server that logs incoming POST requests to the console.

The main purpose of this tool is to act as a webhook endpoint for testing and debugging.

## Features

### Webhook POST Endpoint

The server has an endpoint at `/webhook` that accepts POST requests with JSON payloads.

Upon receiving a POST request, the server will parse the payload as JSON.
If the payload is a valid JSON object, the server will output a log message with the string representation of the object.
If parsing fails, the payload will be logged as invalid JSON and will not be considered further.

### JSON Tracking

The application tracks JSON payloads by a specified key.
For example, consider this sequence of POST payloads:

```json
[
  { "fruit": "apple",  "color": "red" },
  { "fruit": "apple",  "color": "yellow" },
  { "fruit": "banana", "color": "yellow" },
  { "fruit": "cherry", "color": "red" },
  { "fruit": "orange", "color": "orange" }
]
```

If the server is started with the key `fruit`, it will track:

- fruit is apple: 2 times
- fruit is banana: 1 time
- fruit is cherry: 1 time
- fruit is orange: 1 time

If the key were `color`, the application would remember:

- color is red: 2 times
- color is yellow: 2 times
- color is orange: 1 time

### Stats GET Endpoint

The server has a GET endpoint at `/stats` to report the results of the JSON payload tracker.

## Usage

```sh
webhook-logger --help
# Usage of webhook-logger:
#   -k string
#      key used by the JSON payload tracker
#   -p string
#      port where the server listens (default "8080")
```

Open a terminal and run the following command to start the server:

```sh
webhook-logger --port 8080 --key color
# 2026/01/02 15:40:26 Webhook logger server listening on :8080
# 2026/01/02 15:40:26   POST :8080/webhook - receive alerts
# 2026/01/02 15:40:26   GET  :8080/health  - health check
# 2026/01/02 15:40:26   GET  :8080/stats   - view received webhooks
# 2026/01/02 15:40:26 Using key: color
```

From another terminal, you can use `curl` to send POST requests to the webhook endpoint.

```sh
curl -X POST -H "Content-Type: application/json" -d '{"fruit": "apple", "color": "red"}' http://localhost:8080/webhook
```

Upon sending the request, you should see a log message in the server terminal:

```txt
2026/01/02 15:41:07 Received payload: {"fruit": "apple", "color": "red"}
```

And then check the stats:

```sh
curl -s http://localhost:8080/stats | jq
# {
#   "counts": {
#     "red": 1
#   },
#   "key": "color",
#   "total": 1
# }
```

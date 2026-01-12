# Service Monitor

[![Continuous Integration - Go](https://github.com/jzer7/svc-mon/actions/workflows/ci-go.yaml/badge.svg)](https://github.com/jzer7/svc-mon/actions/workflows/ci-go.yaml)
[![license - MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Service Monitor is a tool designed to monitor the health of various services in your infrastructure.
It is opinionated towards simplicity and ease of use, making it easy to configure and deploy.

## Features

- Monitor HTTP/HTTPS services
- Customizable health checks
- Alerts are sent via callbacks
- Simple configuration using YAML files
- Single binary deployment
- Coded in Go for performance and reliability

![Service Monitor deployment](docs/images/deployment-simple.svg)

## Usage

To use Service Monitor, follow these steps:

1. Download the latest release from the [releases page](https://github.com/jzer7/svc-mon/releases).
2. Create a configuration file in YAML format.
   An example configuration is provided in the `examples` directory as `config.yaml`.
3. Run the Service Monitor binary with the configuration file as an argument:

   ```sh
   ./svc-mon --config /path/to/your/config.yaml
   ```

## Configuration

Service Monitor uses a YAML configuration file to define the services to monitor and their health check parameters.
Here is the simplest configuration example:

```yaml
services:
  - name: Example Service
    url: http://example.com/health
    webhook: http://your-webhook-url.com/alert
```

By default, Service Monitor checks each specified URL every 5 minutes with a 5-second timeout.
If a request times out or returns a 5xx HTTP status code, an alert is sent to the configured webhook.
Each webhook receives a JSON payload containing details about the alert.

```json
{
  "service_name": "Example Service",
  "service_url": "http://example.com/health",
  "status": "down",
  "reason": "http_5xx",
  "status_code": 500,
  "timestamp": "2026-01-11T23:17:12Z",
}
```

The payload includes:

- `service_name`: Name of the service that failed
- `service_url`: URL that was monitored
- `status`: Current status (`"up"` or `"down"`)
- `reason`: Failure reason (`"timeout"`, `"http_5xx"`, or `"dns_failure"`)
- `status_code`: HTTP status code (omitted for non-HTTP failures)
- `timestamp`: ISO 8601 timestamp of the check

Advanced configuration options are available. For example:

```yaml
services:
  - name: Example Service
    url: http://service.example.com/health
    interval: 60s
    timeout: 5s
    alert_if:
      - dns_failure
      - timeout
      - http_5xx
    webhooks:
      - http://your-webhook-url.com/alert
      - http://separate-webhook-url.com/alert
defaults:
  interval: 300s
  timeout: 10s
```

The `webhooks` field allows you to specify one or more URLs to receive alerts when a service is down.

## Development

To build Service Monitor from source, ensure you have [Go](https://golang.org/dl/), git and [Task](https://taskfile.dev/) installed.
Then, clone the repository and run the build task:

```sh
git clone https://github.com/jzer7/svc-mon.git
cd svc-mon
task build
```

The compiled binary will be located in the `bin` directory.

To run tests, use the following command:

```sh
task test
```

## License

Service Monitor is licensed under the MIT License.
See the [LICENSE](LICENSE) file for details.

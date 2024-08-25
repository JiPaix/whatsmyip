# whatsmyip

`whatsmyip` is a Go package that provides functionality to determine the external IP address of a machine. It's designed to be reliable, fast, and efficient by using multiple online services concurrently.

## Features

- Fetches the external IP address using multiple online services
- Employs concurrent requests to improve speed and reliability
- Automatically distributes load across different IP lookup services
- Configurable logging based on environment
- Easy to use with a single function call

## Installation

To install the `whatsmyip` package, use the following command:
```console
go get github.com/jipaix/whatsmyip
```

## Usage

Here's a simple example of how to use the `whatsmyip` package:

```go
package main

import (
    "fmt"
    "github.com/jipaix/whatsmyip"
)

func main() {
    ip, url, err := whatsmyip.Get()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Your IP address is: %s\n", ip)
    fmt.Printf("Retrieved from: %s\n", url)
}
```
See more examples in the [/examples](/examples/) folder

## Configuration

The package uses the APP_ENV environment variable to determine the log level:

| APP_ENV | Logging |
| ------- | ------- |
| `local`, `dev`, `development` | Debug level |
| `test`, `staging` | Info level |
| `prod`, `production` | Disabled |
|  Not set | Info level
| Any other value | Disabled  

## How it works

1. The `Get()` function shuffles a list of IP lookup service URLs.
2. It then sends concurrent HTTP GET requests to all URLs.
3. The first successfully retrieved IP address is returned.
4. All ongoing requests are cancelled once a successful response is received.
5. If all requests fail, an error is returned.

## Dependencies

This package depends on the following external library:

- [github.com/charmbracelet/log](https://github.com/charmbracelet/log) for logging

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

see [LICENSE](LICENSE)

## Disclaimer

This package relies on external services to determine the IP address. The availability and accuracy of these services are not guaranteed.

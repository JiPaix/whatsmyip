// Package whatsmyip provides functionality to determine the external IP address of the machine.
//
// This package uses multiple online services to fetch the IP address, improving reliability
// and reducing dependency on any single service. It employs concurrent requests and returns
// the first successful response, cancelling other ongoing requests.
//
// The main function of this package is Get(), which returns the external IP address.
// The package also includes internal utilities for logging.
package whatsmyip

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	l "github.com/charmbracelet/log"
)

// log is the package-level logger instance, configured by the setupLogger function.
// It is used throughout the package for logging debug information and errors.
var log = setupLogger()

// urls is a list of URLs used to fetch the external IP address of the machine.
// These URLs are expected to return a plain/text response containing the IP address.
//
// Expected response formats:
//  1. Single line with IP:
//     "172.201.20.34"
//  2. Single or multiple lines with an "ip=" field:
//     "ip=172.201.20.34"
//
// The order of URLs is randomized before use to distribute load across services.
// This helps prevent overloading any single service with repeated requests.
var urls = []string{
	"https://cloudflare.com/cdn-cgi/trace",
	"https://checkip.amazonaws.com",
	"https://api.ipify.org",
	"https://icanhazip.com",
	"https://myexternalip.com/raw",
	"https://ipinfo.io/ip",
	"https://ipecho.net/plain",
	"https://ifconfig.me/ip",
	"https://ident.me",
	"https://whatismyip.akamai.com",
	"https://wgetip.com",
	"https://ip.tyk.nu",
}

// Get fetches the external IP address of the machine by concurrently querying multiple URLs.
//
// The function performs the following steps:
// 1. Creates a cancellable context
// 2. Shuffles the list of URLs to randomize the order of requests
// 3. Concurrently sends HTTP GET requests to all URLs
// 4. Returns the first successfully retrieved IP address
// 5. Cancels all ongoing requests once a successful response is received
//
// If all requests fail, it returns an error.
//
// Return values:
//   - ip: The retrieved external IP address (empty string if all requests fail)
//   - url: The URL that successfully provided the IP address (empty string if all requests fail)
//   - err: Error if all requests fail, nil otherwise
//
// The function uses the APP_ENV environment variable to determine the log level.
// It logs debug information for successful fetches and an error if all requests fail.
//
// This function is designed to be resilient, fast, and to reduce load on any single IP lookup service.
func Get() (ip string, url string, err error) {
	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan string, len(urls))

	// Shuffle URLs to distribute load across services
	rand.Shuffle(len(urls), func(i, j int) { urls[i], urls[j] = urls[j], urls[i] })

	for _, url := range urls {
		go fetchURL(ctx, url, ch)
	}

	for i := 0; i < len(urls); i++ {
		result := <-ch
		if result != "" {
			cancel() // Cancel other ongoing requests
			log.Debug("Fetch completed", "elapsed", time.Since(start).String(), "pos", i, "url", urls[i])
			return result, urls[i], nil
		}
	}
	return "", "", fmt.Errorf("all requests failed")
}

// fetchURL attempts to retrieve an IP address from the specified URL.
//
// It takes three parameters:
//   - ctx: A context.Context for cancellation and timeouts
//   - url: The URL to fetch the IP address from
//   - ch: A channel to send the result back to the caller
//
// The function performs an HTTP GET request to the given URL. If successful,
// it attempts to extract an IP address from the response body using the getIP function.
// The extracted IP is sent to the channel if successful, otherwise an empty string is sent.
//
// Any error during the process (request creation, HTTP request, body reading, or IP extraction)
// results in an empty string being sent to the channel.
//
// This function is designed to be run as a goroutine in a concurrent fetch operation.
func fetchURL(ctx context.Context, url string, ch chan<- string) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- ""
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- ""
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- ""
		return
	}

	ip, err := parseIP(string(body))
	if err != nil {
		ch <- ""
		return
	}

	ch <- ip
}

// parseIP extracts an IP address from a given string.
//
// Splits the string into lines and tries to parse each line as an IP.
//
// The function is case-insensitive, converting all input to lowercase before processing.
// If a line starts with "ip=", it removes the prefix before parsing.
//
// Parameters:
//   - s: A string containing the potential IP address.
//
// Returns:
//   - string: The extracted IP address.
//   - error: An error if no valid IP address is found.
//
// Error cases:
//   - Returns an error if no valid IP address is found in the input string.
//
// Note: This function does not assume that a single-line response always contains a valid IP address.
// It tries to parse all input as an IP address, regardless of the format.
func parseIP(s string) (string, error) {
	s = strings.ToLower(s)
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		line = strings.TrimPrefix(line, "ip=")
		if netIP := net.ParseIP(line); netIP != nil {
			return netIP.String(), nil
		}
	}
	return "", fmt.Errorf("no ip address found")
}

// setupLogger initializes and returns a configured logger based on the APP_ENV environment variable.
//
// The function sets the log level according to the following APP_ENV values:
//   - "local", "dev", "development": Debug level
//   - "test", "staging": Info level
//   - "prod", "production": Maximum level (effectively disabling logging)
//   - If APP_ENV is not set: Info level
//   - Any other value: Maximum level
//
// The logger is configured with the following options:
//   - Output to stderr
//   - Timestamp reporting enabled
//   - Caller reporting disabled
//   - Time format set to time.DateTime
//   - Prefix set to "🌐 "
//
// Returns:
//   - *github.com/charmbracelet/log.Logger: A configured logger instance
func setupLogger() *l.Logger {
	env, ok := os.LookupEnv("APP_ENV")
	var lvl l.Level
	if !ok {
		lvl = l.InfoLevel
	} else {
		// Set log level based on APP_ENV
		switch strings.ToLower(env) {
		case "local":
			lvl = l.DebugLevel
		case "dev":
			lvl = l.DebugLevel
		case "development":
			lvl = l.DebugLevel
		case "prod":
			lvl = math.MaxInt32 // Effectively disable logging
		case "production":
			lvl = math.MaxInt32 // Effectively disable logging
		case "test":
			lvl = l.InfoLevel
		case "staging":
			lvl = l.InfoLevel
		default:
			lvl = math.MaxInt32 // Effectively disable logging
		}
	}

	return l.NewWithOptions(os.Stderr, l.Options{
		ReportTimestamp: true,
		ReportCaller:    false,
		TimeFormat:      time.DateTime,
		Level:           lvl,
		Prefix:          "🌐 ",
	})
}

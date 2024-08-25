package whatsmyip

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	l "github.com/charmbracelet/log"
)

var log = setupLogger()

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

	ip, err := getIP(string(body))
	if err != nil {
		ch <- ""
		return
	}

	ch <- ip
}

func getIP(s string) (string, error) {
	s = strings.ToLower(s) // Convert to lowercase for case-insensitive matching
	if strings.Contains(s, "\n") {
		lines := strings.Split(s, "\n")

		if len(lines) == 0 {
			log.Error("Empty response")
			return "", fmt.Errorf("empty response")
		}

		if len(lines) == 2 && lines[1] == "" {
			if strings.HasPrefix(lines[0], "ip=") {
				return strings.TrimPrefix(lines[0], "ip="), nil
			} else {
				return lines[0], nil
			}
		}

		for _, line := range lines {
			if strings.HasPrefix(line, "ip=") {
				return strings.TrimPrefix(line, "ip="), nil
			}
		}
	} else {
		return s, nil
	}
	return "", fmt.Errorf("no ip address found")
}

// Get fetches the external IP address of the machine from multiple URLs
// It tries to fetch the IP address from multiple URLs
// and returns the first successful response
// ip returns the public ip
func Get() (ip string, source string, err error) {
	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan string, len(urls))

	// shuffle urls
	rand.Shuffle(len(urls), func(i, j int) { urls[i], urls[j] = urls[j], urls[i] })

	for _, url := range urls {
		go fetchURL(ctx, url, ch)
	}

	for i := 0; i < len(urls); i++ {
		result := <-ch
		if result != "" {
			cancel()
			log.Debug("Fetch completed", "elapsed", time.Since(start).String(), "pos", i, "url", urls[i])
			return result, urls[i], nil
		}
	}
	log.Error("All requests failed")
	return "", "", fmt.Errorf("all requests failed")
}

func setupLogger() *l.Logger {
	env, ok := os.LookupEnv("APP_ENV")
	var lvl l.Level
	if !ok {
		lvl = l.InfoLevel
	} else {
		switch strings.ToLower(env) {
		case "local":
			lvl = l.DebugLevel
		case "dev":
			lvl = l.DebugLevel
		case "development":
			lvl = l.DebugLevel
		case "prod":
			lvl = math.MaxInt32
		case "production":
			lvl = math.MaxInt32
		case "test":
			lvl = l.InfoLevel
		case "staging":
			lvl = l.InfoLevel
		default:
			lvl = math.MaxInt32
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

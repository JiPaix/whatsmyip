package whatsmyip

import (
	"net"
	"os/exec"
	"runtime"
	"slices"
	"testing"
)

var curlSystems = []string{"darwin", "linux", "freebsd"}
var hasCurlEquivalent = []string{"windows"}
var expectedToCURL = slices.Contains(append(curlSystems, hasCurlEquivalent...), runtime.GOOS)
var machineIP string

func init() {
	if slices.Contains(curlSystems, runtime.GOOS) {
		curl := exec.Command("curl", "https://api.ipify.org")
		out, err := curl.Output()
		if err != nil {
			panic(err.Error())
		} else if string(out) == "" {
			panic("empty output from curl")
		} else {
			machineIP = string(out)
		}
	} else if runtime.GOOS == "windows" {
		ipconfig := exec.Command("powershell", "Invoke-WebRequest", " https://api.ipify.org", "|", "Select-Object -Expand Content")
		out, err := ipconfig.Output()
		if err != nil {
			panic(err.Error())
		} else if string(out) == "" {
			panic("empty output from ipconfig")
		} else {
			machineIP = string(out)
		}
	}
}

// Test Get function using the default settings
func TestGet(t *testing.T) {
	ip, source, err := Get()

	if net.ParseIP(ip) == nil {
		t.Errorf("invalid IP address: %s", ip)
	}

	if !slices.Contains(urls, source) {
		t.Errorf("invalid source: %s", source)
	}

	if err != nil {
		t.Errorf("error: %s", err)
	}

	// We're only testing this with OS w/ curl or equivalent
	if expectedToCURL {
		if ip != machineIP {
			t.Errorf("expected %s, got %s", machineIP, ip)
		}
	}
}

// Test Get function by tempering the url list with a bad URL
func TestGetWithBadURL(t *testing.T) {
	urls = []string{"https://example.org"}

	ip, source, err := Get()

	if net.ParseIP(ip) != nil {
		t.Errorf("found address: %s", net.ParseIP(ip).String())
	}

	if source != "" {
		t.Errorf("found source: %s", source)
	}

	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

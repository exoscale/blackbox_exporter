package prober

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	// "os"
	"strings"
	"time"

	"github.com/prometheus/blackbox_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

func extractVMCURL(txt string) string {
	if !strings.HasPrefix(txt, "v=BIMI1") {
		return ""
	}

	parts := strings.Split(txt, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "l=") {
			return strings.TrimPrefix(part, "l=")
		}
	}
	return ""
}

func ProbeVMC(ctx context.Context, target string, module config.Module, registry *prometheus.Registry, logger *slog.Logger) error {
	baseDomain := "exoscale.com"
	bimiDomain := "default._bimi." + baseDomain
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}

	txtRecords, err := resolver.LookupTXT(ctx, bimiDomain)
	if err != nil {
		fmt.Printf("Error looking up TXT record for %s: %v\n", bimiDomain, err)
		return err
	}

	found := false
	for _, txt := range txtRecords {
		fmt.Println("Found TXT record:", txt)
		if strings.HasPrefix(txt, "v=BIMI1") {
			parts := strings.Split(txt, ";")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "l=") {
					url := strings.TrimPrefix(part, "l=")
					fmt.Println("Found URL:", url)
					found = true
					break
				}
			}
		}
	}

	if !found {
		fmt.Println("No valid BIMI record found.")
	}

	return nil
}

func ProbeVMCWrapper(ctx context.Context, target string, module config.Module, registry *prometheus.Registry) bool {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	err := ProbeVMC(ctx, target, module, registry, logger)
	if err != nil {
		fmt.Printf("ProbeVMC failed: %v\n", err)
		return false
	}
	return true
}

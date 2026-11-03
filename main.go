// tynet-ddns keeps one Cloudflare DNS A record pointed at the current
// public IP. Designed to run as a Kubernetes CronJob — one check-and-
// maybe-update pass per invocation, no internal scheduling loop. A
// non-zero exit surfaces as a failed Job; the next scheduled run just
// tries again.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "log the diff but don't update the record")
	flag.Parse()

	if err := run(*dryRun); err != nil {
		log.Fatal(err)
	}
}

func run(dryRun bool) error {
	token := os.Getenv("CF_API_TOKEN")
	zoneID := os.Getenv("CF_ZONE_ID")
	recordID := os.Getenv("CF_RECORD_ID")

	for name, val := range map[string]string{
		"CF_API_TOKEN": token,
		"CF_ZONE_ID":   zoneID,
		"CF_RECORD_ID": recordID,
	} {
		if val == "" {
			return fmt.Errorf("%s env var is required", name)
		}
	}

	httpClient := http.DefaultClient

	currentIP, err := CurrentPublicIP(httpClient)
	if err != nil {
		return fmt.Errorf("getting current public IP: %w", err)
	}

	cf := &CloudflareClient{
		HTTPClient: httpClient,
		Token:      token,
		ZoneID:     zoneID,
		RecordID:   recordID,
	}

	recordIP, err := cf.CurrentContent()
	if err != nil {
		return fmt.Errorf("getting current record content: %w", err)
	}

	if recordIP == currentIP {
		log.Printf("no change: record already points at %s", currentIP)
		return nil
	}

	if dryRun {
		log.Printf("dry-run: would update record from %s to %s", recordIP, currentIP)
		return nil
	}

	if err := cf.UpdateContent(currentIP); err != nil {
		return fmt.Errorf("updating record: %w", err)
	}

	log.Printf("updated record: %s -> %s", recordIP, currentIP)
	return nil
}

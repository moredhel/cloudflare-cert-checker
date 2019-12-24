package main

// TODO: fix up issue regarding long requests being hardcoded at a max of 5 seconds

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/genkiroid/cert"
	"github.com/dustin/go-humanize"
)


type Cert struct {
	Cert cert.Cert
	Domain string
}

type CertResult struct {
	Domain string
	Error string
	NotAfter time.Time
}

func (cert *CertResult) String() string {
	val := humanize.Time(cert.NotAfter)
	if cert.Error != "" {
		val = cert.Error
	}
	return fmt.Sprintf("%s: %s", cert.Domain, val)
}

func getDomainCert(domains []string, ret chan Cert) {
	certs, err := cert.NewCerts(domains)
	if err != nil {
		log.Fatal(err)
	}

	for _, cert := range certs {
		c := Cert{
			Domain: cert.DomainName,
			Cert: *cert,
		}
		ret <- c
	}
}

func getExpiry(cert Cert) CertResult {
	results := CertResult{
		Domain: cert.Domain,
	}

	t, err := time.Parse("2006-01-02 15:04:05 -0700 CET", cert.Cert.NotAfter)
	if err != nil {
		log.Printf("fail: %s\n", cert.Domain)
		results.Error = fmt.Sprintf("fail: %s", cert.Cert)
	}
	results.NotAfter = t

	return results
}

// Handlers
func handleZone(api *cloudflare.API, zone cloudflare.Zone, ret chan Cert, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("'%s' zone starting\n", zone.Name)
	// get Domain records
	recs, err := api.DNSRecords(zone.ID, cloudflare.DNSRecord{})
	if err != nil {
		fmt.Println(err)
		return
	}

	names := []string{}
	for _, domain := range recs {
		ty := domain.Type
		if ty == "A" || ty == "CNAME" {
			names = append(names, domain.Name)
		}
	}

	getDomainCert(names, ret)
	log.Printf("'%s' done\n", zone.Name)
}

func handleCert(certChan chan Cert, wg *sync.WaitGroup) {
	for {
		select {
		case cert := <-certChan:
			expire := getExpiry(cert)
			fmt.Printf("%s\n",expire.String())
		case <-time.After(5*time.Second): // TODO: improve this...
			wg.Done()
		}
	}
}

func main() {
	api, err := cloudflare.New(os.Getenv("CLOUDFLARE_API_KEY"), os.Getenv("CLOUDFLARE_EMAIL"))
	if err != nil {
		log.Fatal(err)
	}

	// Fetch the zones
	zones, err := api.ListZones()
	if err != nil {
		log.Fatal(err)
	}

	certChan := make(chan Cert)
	var wg sync.WaitGroup
	// Fetch domains from each zone
	for _, zone := range zones {
		wg.Add(1)
		go handleZone(api, zone, certChan, &wg)
		// break
	}

	wg.Add(1)
	// process returned certs
	log.Printf("dispatching cert handler...")
	go handleCert(certChan, &wg)
	wg.Wait()
	log.Printf("closing...")
}

package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/btcsuite/btcrpcclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultAddress = "127.0.0.1:8332"
	defaultListen  = "127.0.0.1:9099"
)

var version = "v0.1.0-dev"

var (
	address  string
	username string
	password string
	listen   string
)

var (
	BlockCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "earthcoin",
			Name:      "block_count_total",
			Help:      "Total number of Blocks",
		},
	)
	Subsidy = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "earthcoin",
			Name:      "subsidy",
			Help:      "Current subsidy",
		},
	)
	ConnectionCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "earthcoin",
			Name:      "connection_count",
			Help:      "Current number of connections",
		},
	)
	Difficulty = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "earthcoin",
			Name:      "difficulty",
			Help:      "Current difficulty",
		},
	)
	HashesPerSec = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "earthcoin",
			Name:      "hashes_per_second",
			Help:      "Current hashes per second",
		},
	)
	NetworkHashesPerSec = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "earthcoin",
			Name:      "network_hashes_per_second",
			Help:      "Current network hashes per second",
		},
	)
)

func init() {
	kingpin.Flag("address", "Address of the JSON-RPC endpoint to conenct to").Default(defaultAddress).StringVar(&address)
	kingpin.Flag("username", "The username to use when connecting to the endpoint").StringVar(&username)
	kingpin.Flag("password", "The password to use when connecting to the endpoint").Envar("COIN_EXPORTER_PASSWORD").StringVar(&password)
	kingpin.Flag("listen", "Address to serve the collected metrics on").Default(defaultListen).StringVar(&listen)

	prometheus.MustRegister(
		BlockCount,
		Subsidy,
		ConnectionCount,
		Difficulty,
		HashesPerSec,
		NetworkHashesPerSec,
	)
}

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &btcrpcclient.ConnConfig{
		Host:         address,
		User:         username,
		Pass:         password,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := btcrpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	go retrieveMetrics(client)

	serveMetrics(listen)
}

func retrieveMetrics(client *btcrpcclient.Client) {
	for {
		// Get the current block count.
		blockCount, err := client.GetBlockCount()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Block count: %d", blockCount)

		// Export current block count as prometheus metric.
		BlockCount.Set(float64(blockCount))

		bestBlockHash, err := client.GetBlockHash(blockCount)
		if err != nil {
			log.Fatal(err)
		}

		bestBlock, err := client.GetBlock(bestBlockHash)
		if err != nil {
			log.Fatal(err)
		}

		subsidy := bestBlock.Transactions[0].TxOut[0].Value / 100000000
		log.Printf("Subsidy: %d", subsidy)

		// TODO
		Subsidy.Set(float64(subsidy))

		// Get the current connection count.
		connectionCount, err := client.GetConnectionCount()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Connection count: %d", connectionCount)

		// Export current connection count as prometheus metric.
		ConnectionCount.Set(float64(connectionCount))

		// Get the current difficulty.
		difficulty, err := client.GetDifficulty()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Difficulty: %f", difficulty)

		// Export current difficulty as prometheus metric.
		Difficulty.Set(difficulty)

		// Get the current hashes per second.
		hashesPerSec, err := client.GetHashesPerSec()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Hashes per second: %d", hashesPerSec)

		// Export current hashes per second as prometheus metric.
		HashesPerSec.Set(float64(hashesPerSec))

		// Get the current network hashes per second.
		networkHashesPerSec, err := client.GetNetworkHashPS()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Network hashes per second: %d", networkHashesPerSec)

		// Export current network hashes per second as prometheus metric.
		NetworkHashesPerSec.Set(float64(networkHashesPerSec))

		// sleeping before next cycle
		time.Sleep(time.Minute)
	}
}

func serveMetrics(address string) {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(address, nil))
}

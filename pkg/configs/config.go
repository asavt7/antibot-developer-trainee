package configs

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	port            int
	prefixSize      int
	requestLimit    int
	interval        time.Duration
	blockingTimeout time.Duration
)

func init() {
	const (
		defaultPort            = 8080
		defaultPrefixSize      = 24
		defaultRequestLimit    = 10
		defaultTimeLimit       = 10 * time.Second
		defaultBlockingTimeout = 100 * time.Second
	)
	flag.IntVar(&port, "port", lookupEnvOrInt("PORT", defaultPort), "port number")
	flag.IntVar(&prefixSize, "length", lookupEnvOrInt("LENGTH", defaultPrefixSize), "subnet prefix length [0..32]")
	flag.IntVar(&requestLimit, "limit", lookupEnvOrInt("LIMIT", defaultRequestLimit), "maximum number of requests per interval ${interval}")
	flag.DurationVar(&interval, "interval", lookupEnvOrDuration("INTERVAL", defaultTimeLimit), "interval")
	flag.DurationVar(&blockingTimeout, "blocking_timeout", lookupEnvOrDuration("BLOCKING_TIMEOUT", defaultBlockingTimeout), "resource blocking time if request quota is exceeded")
}

func lookupEnvOrDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		duration, err := time.ParseDuration(val)
		if err != nil {
			log.Fatalf("illegal value for ENV %s: %v", key, err)
		}
		return duration
	}
	return defaultVal
}

func lookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("illegal value for ENV %s: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func NewConfigs() Config {
	flag.Parse()
	validateCLIArgs()
	c := Config{
		Port:            port,
		PrefixSize:      prefixSize,
		RequestLimit:    requestLimit,
		TimeInterval:    interval,
		BlockingTimeout: blockingTimeout,
	}
	log.Printf("Configuration: %+v", c)
	return c
}

func validateCLIArgs() {
	if prefixSize < 0 || prefixSize > 32 {
		log.Fatalf("Illegal argument subnet prefix length!")
	}
}

type Config struct {
	Port int

	PrefixSize      int
	RequestLimit    int
	TimeInterval    time.Duration
	BlockingTimeout time.Duration
}

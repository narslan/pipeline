package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/narslan/dataflow/cassandra"
	"github.com/narslan/dataflow/pipeline"
	"github.com/narslan/dataflow/redis"
	"github.com/narslan/dataflow/s3"
	"github.com/BurntSushi/toml"
)

// main is the entry point into our application. It doesn't return errors.
// Because of that we delegate our program to the Run() function.
func main() {
	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	// Instantiate a new type to represent our application.
	m := NewMain()

	// Parse command line flags & load configuration.
	err := m.ParseFlags(ctx, os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	errCh := make(chan error)

	// Running on a new goroutine.
	go m.Run(ctx, errCh)

	// Listen on error and interruption.
	select {
	case err := <-errCh: // errCh receives nil if pipeline succefully finishes.
		if err != nil {
			// we can gracefully print error.
			fmt.Println(err)
		}
	case <-c:
		fmt.Println("closing")
		// Release resources.
		m.Close()
	case <-ctx.Done():
		fmt.Println("caught interrupt")
		m.Close()
	}

}

func (m *Main) ParseFlags(ctx context.Context, args []string) error {
	// Our flag set is very simple. It only includes a config path and concurrency parameter.
	//It fails if config file is not supplied.
	flag.StringVar(&m.ConfigPath, "config", "", "config path")
	flag.IntVar(&m.NumCPU, "concurrency", 0, "number of concurrent goroutines")

	// Custom error handling
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), "Supply a config file similar to:\n")
		fmt.Printf("%s  -config path -concurrency 4\n   ", os.Args[0])
	}
	flag.Parse()

	if m.ConfigPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if m.NumCPU == 0 {
		m.NumCPU = runtime.NumCPU()
	}

	// Read our TOML formatted configuration file.
	config, err := ReadConfigFile(m.ConfigPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", m.ConfigPath)
	} else if err != nil {
		return err
	}
	m.Config = config

	return nil
}

// Main represents the program with .
type Main struct {
	// Configuration path and parsed config data.
	Config     Config
	ConfigPath string
	NumCPU     int
	DB         *cassandra.DB
}

// NewMain returns a new instance of Main.
func NewMain() *Main {
	return &Main{

		// This field is here to make Close method of the DB available for Main.
		DB: &cassandra.DB{},
	}
}

// Close gracefully closes db connection.
func (m *Main) Close() error {

	if m.DB != nil {
		m.DB.Close()
	}

	return nil
}

type Config struct {
	Cassandra struct {
		Host     string `toml:"host"`
		Keyspace string `toml:"keyspace"`
		User     string `toml:"user"`
		Pass     string `toml:"pass"`
	} `toml:"cassandra"`

	Redis struct {
		Addr string `toml:"addr"`
		Pass string `toml:"pass"`
		DB   int    `toml:"db"`
	} `toml:"redis"`
}

// ReadConfigFile unmarshals config from file.
func ReadConfigFile(filename string) (Config, error) {
	config := Config{}
	if buf, err := os.ReadFile(filename); err != nil {
		return config, err
	} else if err := toml.Unmarshal(buf, &config); err != nil {
		return config, err
	}
	return config, nil
}

// Run executes the main program, which starts the pipeline job.
func (m *Main) Run(ctx context.Context, errCh chan<- error) {

	defer timeTrack(time.Now(), "Pipeline")

	fmt.Println("Executing Pipeline")
	// Name of the files, that are on S3. They will get by S3Service.
	files := []string{
		"products-1.jsonl",
		"products-2.jsonl",
		"products-3.jsonl",
		"products-4.jsonl"}

	// Set Cassandra connection params that comes from config file.
	dbhost := m.Config.Cassandra.Host
	keyspace := m.Config.Cassandra.Keyspace
	user := m.Config.Cassandra.User
	pass := m.Config.Cassandra.Pass

	// Connect to Cassandra.
	fmt.Println("Connecting Cassandra")
	db, err := cassandra.NewDB(dbhost, keyspace, user, pass)
	if err != nil {
		errCh <- err
	}

	fmt.Println("Succefuly connected to Cassandra")
	// Assign to the DB instance of main program, to be able to close database from the main program.
	m.DB = db

	// Instantiate Cassandra-backed service. This service manages operations on Cassandra.
	productService := cassandra.NewProductService(m.DB)

	// Set Redis connection params from config file.
	addr := m.Config.Redis.Addr
	rpass := m.Config.Redis.Pass
	rdb := m.Config.Redis.DB

	// Connect to Redis.
	fmt.Println("Connecting to Redis")
	cache, err := redis.NewCache(addr, rpass, rdb)
	if err != nil {
		errCh <- err
	}

	fmt.Println("Connected to Redis")
	// Instantiate Redis-backed cache service. This service manages operations on the database.
	cacheService := redis.NewIDCacheService(cache)

	// Start S3 service
	bucket, ok := os.LookupEnv("AWS_S3_BUCKET")
	if !ok {
		errCh <- errors.New("AWS_S3_BUCKET environment variable not found")
	}

	// Instantiate S3 fetcher, which retrieves file from S3.
	s3Service, err := s3.NewS3FetchService(bucket)
	if err != nil {
		errCh <- err
	}

	// Make a pipeline from s3Service and key names.
	pipe := pipeline.NewPipeline(s3Service, m.NumCPU)

	// Bind services to the pipeline.
	pipe.ProductService = productService
	pipe.CacheService = cacheService

	// Read JSONL files and return them as a channel of byte slice.
	fmt.Println("Starting pipeline")
	// Kick start the pipeline. Send the error or completion result in the errCh
	errCh <- pipe.Run(context.TODO(), files...)

}

// timeTrack tells how long it takes for a function to run.
// Adapted from https://blog.stathat.com/2012/10/10/time_any_function_in_go.html
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s tooks %s", name, elapsed)
}

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/narslan/dataflow/cassandra"
	"github.com/narslan/dataflow/http"
	"github.com/BurntSushi/toml"
)

// main is the entry point to our application.
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

	// Execute program.
	if err := m.Run(ctx); err != nil {
		m.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	//Wait for ctrl-c
	<-ctx.Done()

	// Clean up program.
	if err := m.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Close gracefully stops the program.
func (m *Main) Close() error {
	if m.HTTPServer != nil {
		if err := m.HTTPServer.Close(); err != nil {
			return err
		}
	}

	if m.DB != nil {
		m.DB.Close()
	}

	return nil
}

func (m *Main) ParseFlags(ctx context.Context, args []string) error {
	// Our flag set is very simple. It only includes a config path.
	//It fails if config file is not supplied.

	flag.StringVar(&m.ConfigPath, "config", "", "config path")
	// Custom error handling
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), "Supply a config file similar to:\n")
		fmt.Printf("%s  -config path\n   ", os.Args[0])
	}
	flag.Parse()

	if m.ConfigPath == "" {
		flag.Usage()
		os.Exit(1)
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

	DB         *cassandra.DB
	HTTPServer *http.Server
}

// NewMain returns a new instance of Main.
func NewMain() *Main {
	return &Main{
		Config: DefaultConfig(),

		// This is here to make Close method of the DB available for Main, not beautiful to have here.
		DB:         &cassandra.DB{},
		HTTPServer: http.NewServer(),
	}
}

type Config struct {
	HTTP struct {
		Address string `toml:"address"`
		Domain  string `toml:"domain"`
	} `toml:"http"`

	Cassandra struct {
		Host     string `toml:"host"`
		Keyspace string `toml:"keyspace"`
		User     string `toml:"user"`
		Pass     string `toml:"pass"`
	} `toml:"cassandra"`
}

// DefaultConfig returns a new instance of Config with defaults set.
func DefaultConfig() Config {
	var config Config
	return config
}

// ReadConfigFile unmarshals config from file.
func ReadConfigFile(filename string) (Config, error) {
	config := DefaultConfig()
	if buf, err := os.ReadFile(filename); err != nil {
		return config, err
	} else if err := toml.Unmarshal(buf, &config); err != nil {
		return config, err
	}
	return config, nil
}

// Run executes the program.
func (m *Main) Run(ctx context.Context) error {

	// Set Cassandra connection params from config file.
	dbhost := m.Config.Cassandra.Host
	keyspace := m.Config.Cassandra.Keyspace
	user := m.Config.Cassandra.User
	pass := m.Config.Cassandra.Pass

	// Create a
	db, err := cassandra.NewDB(dbhost, keyspace, user, pass)
	if err != nil {
		return err
	}
	// Assign to the DB instance of main program.
	// This is required, as main program should be able to close the db instance.
	m.DB = db
	// Instantiate Cassandra-backed service.
	productService := cassandra.NewProductService(m.DB)

	m.HTTPServer.Address = m.Config.HTTP.Address
	// Attach underlying services to the HTTP server.
	m.HTTPServer.ProductService = productService

	// Start the HTTP server.
	return m.HTTPServer.Open()

	//TODO: Enable internal debug endpoints.

}

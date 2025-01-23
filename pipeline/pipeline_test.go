package pipeline_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"bitbucket.org/nevroz/dataflow"
	"bitbucket.org/nevroz/dataflow/cassandra"
	"bitbucket.org/nevroz/dataflow/container"
	"bitbucket.org/nevroz/dataflow/pipeline"
	"bitbucket.org/nevroz/dataflow/redis"
)

// FileReader represents a file reader for local filesytem. It mimicks the S3FetchService.
// It is used just for testing.
type FileReader struct{}

// Ensure that S3FetchService implements dataflow.Fetch.
var _ dataflow.Fetch = (*FileReader)(nil)

// Get method reads a local file and return data of the file as byte slice.
func (s *FileReader) Get(ctx context.Context, path string) ([]byte, error) {
	return os.ReadFile(path)
}

// Assign how many worker goroutines for the pipeline should use.
var n = runtime.NumCPU()

func TestLoadFiles(t *testing.T) {
	//Ensure that file sources loaded into pipeline.

	// Find the paths of all input files in the data directory.
	paths, err := filepath.Glob(filepath.Join("testdata", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}

	// Assign how many worker goroutines for the pipeline.
	n := runtime.NumCPU()
	//Testing first stage of the pipeline with single files.
	t.Run("OK", func(t *testing.T) {

		for _, path := range paths {
			t.Run(path, func(t *testing.T) {

				// The pipeline needs a fetcher.
				f := &FileReader{}

				// Provide reader and the name of the file to the pipeline.
				pipe := pipeline.NewPipeline(f, n)

				// Read JSONL files and return them as a channel of byte slice.
				fileCh, _, err := pipe.LoadFiles(context.TODO(), path)
				if err != nil {
					t.Fatal(err)
				}

				// Step through the channel, capture the content of the files.
				for got := range fileCh {

					// Load local file for comparison.
					want, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(f)
					}

					if !bytes.Equal(got, want) {
						t.Fatalf("%q: content mismatch in path", path)
					}
				}
			})
		}
	})

	t.Run("ErrNoSource", func(t *testing.T) {
		// Ensure that without a input source error returns.
		// Assign how many worker goroutines for the pipeline.
		n := runtime.NumCPU()
		f := &FileReader{}

		// Provide reader and no file.
		// Provide reader and the name of the file to the pipeline.
		pipe := pipeline.NewPipeline(f, n)

		// Read JSONL files and return them as a channel of byte slice.
		_, _, err := pipe.LoadFiles(context.TODO())
		if err == nil {
			t.Fatal(err)
		}

		if err.Error() != "no sources provided" {
			t.Fatalf("expected an error %v, got %v", err, err)
		}
	})

}

func TestSplitFiles(t *testing.T) {
	// Ensure that files are splitted by newline.

	// Test case.
	// Find the paths of all input files in the data directory.
	paths, err := filepath.Glob(filepath.Join("testdata", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}

	// Those files has 5000 lines in total. The test will be asking if it is true.
	want := 5000

	// The pipeline needs a fetcher.
	f := &FileReader{}

	// Context for pipeline methods.
	ctx := context.Background()

	// Provide reader and the name of the file to the pipeline.
	pipe := pipeline.NewPipeline(f, n)

	// Read JSONL files and return them as a channel of byte slice.
	fileCh, _, err := pipe.LoadFiles(context.TODO(), paths...)

	if err != nil {
		t.Fatal(err)
	}
	// Split JSONL data at new lines.
	// Return a string channel.
	// Send JSON strings through it.
	linesCh, _ := pipe.Split(ctx, fileCh)

	// Count the lines flowing through the channel.
	var got int
	for range linesCh {
		got++
	}

	if got != want {
		t.Fatalf("content mismatch; expected %d, got %d", want, got)
	}
}

func TestConvertLine(t *testing.T) {
	// Ensure that json lines are converted into Product entities.

	// Test case.
	// Find the paths of all input files in the data directory.
	paths, err := filepath.Glob(filepath.Join("testdata", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}

	// Those files has 5000 lines in total. The test will be asking if it is true.
	want := 5000

	// The pipeline needs a fetcher.
	f := &FileReader{}
	// Context for pipeline methods.
	ctx := context.Background()

	// Provide reader and the path of the files to the pipeline.
	pipe := pipeline.NewPipeline(f, n)

	// Read JSONL files and return them as a channel of byte slice.
	fileCh, _, err := pipe.LoadFiles(context.TODO(), paths...)

	if err != nil {
		t.Fatal(err)
	}

	// Split JSONL data at new lines.
	linesCh, _ := pipe.Split(ctx, fileCh)

	// Convert JSON strings into Product types.
	convertCh, _ := pipe.ConvertJSON(ctx, linesCh)

	// A container for the output of the pipeline.
	products := make([]*dataflow.Product, 0)
	for v := range convertCh {
		products = append(products, v)
	}

	got := len(products)
	if got != want {
		t.Fatalf("content mismatch; expected %d, got %d", want, got)
	}

}

func TestRunPipeline(t *testing.T) {
	//Ensure that Products are saved.

	// 	// context for containers.
	ctx := context.Background()

	// Start containers for test.
	cdbc, cassandraConnectionHost := container.MustDeployCassandra(ctx)
	defer container.MustCleanCassandraContainer(ctx, cdbc)

	rdbc, redisConnectionString := container.MustDeployRedis(ctx)
	defer container.MustCleanRedisContainer(ctx, rdbc)

	// Open
	db := MustOpenDB(t, cassandraConnectionHost)
	defer MustCloseDB(t, db)

	// Create a new redis instance
	cache := MustOpenCache(t, redisConnectionString)
	defer MustCloseCache(t, cache)

	// Find the paths of all input files in the data directory.
	paths, err := filepath.Glob(filepath.Join("testdata", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}

	// Pass cassandra db instance to the ProductService.
	s := cassandra.NewProductService(db)

	// Pass redis instance to the IDCacheService.
	idcs := redis.NewIDCacheService(cache)

	// Make an instance of FileReader.
	f := &FileReader{}

	// Make a pipeline from file reader and pathnames.
	pipe := pipeline.NewPipeline(f, n)

	// Bind services to the pipeline.
	pipe.ProductService = s
	pipe.CacheService = idcs

	// Kick start the pipeline.
	err = pipe.Run(context.TODO(), paths...)
	if err != nil {
		t.Fatal(err)
	}

	// ID and Title of the first entry in the file. Check if they are in the database.
	want := struct {
		ID    uint32
		Title string
	}{
		ID:    151000,
		Title: "title151000",
	}

	// Look up in database.
	got, err := pipe.ProductService.FindProductByID(context.TODO(), want.ID)

	// If the product is not found, it is an error, test fails here.
	if err != nil {
		if dataflow.ErrorCode(err) != dataflow.ENOTFOUND {
			t.Fatalf("content mismatch; expected %s, got %s", want.Title, got.Title)
		} else {
			t.Fatal(err)
		}

	}

}

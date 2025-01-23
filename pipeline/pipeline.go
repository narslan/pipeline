package pipeline

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"bitbucket.org/nevroz/dataflow"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Pipeline represents a data flow architecture, which takes some input source
// process it and save into a database.
type Pipeline struct {
	Fetcher dataflow.Fetch //Fetcher is an instance of remote or local services that fetches data.

	NumThreads int // Number of worker threads for LoadFiles and Save methods.

	// Services used by the pipeline.
	// ProductService is used by the Save method.
	ProductService dataflow.ProductService

	// CacheService is also used by the Save method.
	CacheService dataflow.Cache
}

func NewPipeline(f dataflow.Fetch, num int) *Pipeline {

	return &Pipeline{Fetcher: f, NumThreads: num}
}

// LoadFiles step through a list of data concurrently.
// Send the data in to a channel of bytes and return an error channel
// and error for erros outside the goroutine.
func (p *Pipeline) LoadFiles(ctx context.Context, keys ...string) (<-chan []byte, <-chan error, error) {

	// Fail if no source provided.
	if len(keys) == 0 {
		return nil, nil, errors.New("no sources provided")
	}
	outCh := make(chan []byte)
	errCh := make(chan error, 1)

	// Create a semaphore. A semaphore limits the number of concurrent executions.
	sem := semaphore.NewWeighted(int64(p.NumThreads))

	go func() {
		defer close(outCh)
		var g errgroup.Group
		for _, path := range keys {
			// Acquire semaphore before starting goroutine
			if err := sem.Acquire(ctx, 1); err != nil {
				errCh <- err
				return
			}
			// Start goroutine for fetching data
			g.Go(func() error {
				defer sem.Release(1) // Release semaphore when goroutine finishes
				data, err := p.Fetcher.Get(ctx, path)
				if err != nil {
					errCh <- err
				}
				outCh <- data
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			errCh <- err
		}
	}()
	return outCh, errCh, nil

}

// Split takes slice of JSONL bytes and split them at newlines.
// Send them through the channel.
func (p *Pipeline) Split(ctx context.Context, input <-chan []byte) (<-chan string, <-chan error) {
	outCh := make(chan string)
	errCh := make(chan error, 1)
	go func() {
		defer fmt.Println("Finished splitting")
		defer close(outCh)

		// For each input source data
		for data := range input { // Read from the channel

			// Create a scanner job.
			scanner := bufio.NewScanner(bytes.NewReader(data))
			// Step through bytes line by line.
			for scanner.Scan() {

				// We use Text method instead of Bytes. Explanation is in the link.
				// https://github.com/golang/go/issues/35725#issuecomment-556936725

				outCh <- scanner.Text()
			}
		}
	}()
	return outCh, errCh

}

// Convert takes a JSON line and converts it to Product type.
func (p *Pipeline) ConvertJSON(ctx context.Context, input <-chan string) (<-chan *dataflow.Product, <-chan error) {
	outCh := make(chan *dataflow.Product)
	errCh := make(chan error, 1)
	go func() {
		defer close(outCh)
		for job := range input { // Read from the channel
			var p dataflow.Product
			err := json.Unmarshal([]byte(job), &p)
			if err != nil {
				errCh <- err
			}
			outCh <- &p
		}
	}()
	return outCh, errCh

}

// Save setups a concurrent pipeline stage that calls SendToDB method.
func (p *Pipeline) Save(ctx context.Context, products <-chan *dataflow.Product) (<-chan error, error) {
	// Create a semaphore to limit concurrent executions
	sem := semaphore.NewWeighted(int64(p.NumThreads))

	errc := make(chan error, 1)
	go func() {
		defer close(errc)

		var g errgroup.Group

		for pr := range products {
			// Acquire semaphore before starting goroutine
			if err := sem.Acquire(ctx, 1); err != nil {
				errc <- err
			}
			// Start goroutine for each product
			g.Go(func(pr *dataflow.Product) func() error {
				return func() error {
					defer sem.Release(1) // Release semaphore when goroutine finishes
					select {
					case errc <- p.SendToDB(ctx, pr):
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}(pr))
		}

		// Wait for all goroutines to finish
		if err := g.Wait(); err != nil {
			errc <- err
		}
	}()
	return errc, nil
}

// SendToDB sends a Product type to a database.
// It checks the cache first, looking up for the product ID.
// If the ID already is in the cache, it will not visit database anymore.
// It is called by the method Save.
func (p *Pipeline) SendToDB(ctx context.Context, pr *dataflow.Product) error {

	// Check cache if the ID is there.
	ok, err := p.CacheService.Exists(context.TODO(), pr.ID)
	if err != nil {
		return err
	}

	// If the ID exists in the cache do nothing, just return.
	if ok {
		return nil
	}

	// If the ID does not exist, save product in the DB.
	err = p.ProductService.CreateProduct(context.TODO(), pr)
	if err != nil {
		return err
	}

	// Save the ID it in the cache.
	return p.CacheService.Set(context.TODO(), pr.ID)

}

// Run setups and executes the pipeline. It constructs a list error channels out of
// pipeline stage methods (LoadFiles, Split, ConvertJSON). After that it waits their executions.

func (p *Pipeline) Run(ctx context.Context, paths ...string) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errcList := make([]<-chan error, 0)
	// Source pipeline stage.
	fileCh, errc, err := p.LoadFiles(ctx, paths...)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)
	// Transformer pipeline stage.
	lineCh, errc := p.Split(ctx, fileCh)
	errcList = append(errcList, errc)

	productCh, errc := p.ConvertJSON(ctx, lineCh)
	errcList = append(errcList, errc)

	// This stage save Products into the DB and Cache.
	errc, err = p.Save(ctx, productCh)
	if err != nil {
		return err
	}

	errcList = append(errcList, errc)
	fmt.Println("Pipeline started. Waiting for pipeline to complete.")
	fmt.Printf("Pipeline uses %d number of goroutines for interaction with DB\n", p.NumThreads)
	return wait(ctx, errcList...)
}

// merge merges a number of error channel, and merges into one channel.
func merge(ctx context.Context, cs ...<-chan error) <-chan error {
	out := make(chan error)
	g, ctx := errgroup.WithContext(ctx)

	// Start goroutines for each error channel..
	for _, c := range cs {
		g.Go(func() error {
			defer close(out)
			for err := range c {
				select {
				case out <- err:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	// Wait for all goroutines to finish.
	go func() {
		if err := g.Wait(); err != nil {
			out <- err
		}
	}()

	return out
}

// wait send channels to merge function and wait on returned channel.
func wait(ctx context.Context, errcList ...<-chan error) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Merge error channels
	errc := merge(ctx, errcList...)
	for {
		select {
		// Check if merged error channel closed.
		case err, ok := <-errc:
			if !ok {
				// All channels are closed, no more errors.
				return nil
			}
			if err != nil {
				return err // Return on the first error.
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

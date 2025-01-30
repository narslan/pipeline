package s3_test

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/narslan/dataflow/s3"
)

// If dump flag set true, developer can get files from S3 and save to local file for further examination.
var dump = flag.Bool("dump", false, "save work data")

// MustConnect checks existence of the required environment variables.
// and the availability of s3 service.
func MustConnect(tb testing.TB) (*s3.S3FetchService, error) {

	tb.Helper()

	bucket, ok := os.LookupEnv("AWS_S3_BUCKET")
	if !ok {
		return nil, errors.New("AWS_CASE_BUCKET environment variable not found")
	}

	return s3.NewS3FetchService(bucket)
}

// TestGetSave downloads files for testing purposes.
func TestGetSave(t *testing.T) {

	if !*dump {
		t.Skip("skipping as dump mode set false")
	}
	//Create a S3 service.
	s, err := MustConnect(t)

	if err != nil {
		t.Fatal(err)
	}

	//Create testdata directory where files will be saved.
	dir := "testdata"
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	// Four fiels are available to us.
	testCases := []struct {
		f string
	}{
		{
			f: "products-1.jsonl",
		},
		{
			f: "products-2.jsonl",
		},
		{
			f: "products-3.jsonl",
		},
		{
			f: "products-4.jsonl",
		},
	}

	for _, tC := range testCases {

		// Strip extension for a better test name.
		testName := strings.TrimSuffix(tC.f, filepath.Ext(tC.f))

		// Construct file path from dir and filename.

		filePath := filepath.Join(dir, tC.f)
		//Retrieve files and save under testdata file
		t.Run(testName, func(t *testing.T) {

			// Skip if they are allready there.
			if fileExists(filePath) {
				return
			}

			//Get the resource from AWS S3.
			body, err := s.Get(context.Background(), tC.f)
			if err != nil {
				t.Fatal(err)
			}

			//construct the filename  "path/key".
			filename := filepath.Join(filePath, tC.f)

			f, err := os.Create(filename)

			if err != nil {
				t.Fatal(err)
			}

			//save file under path.
			_, err = f.Write(body)
			if err != nil {
				t.Fatal(err)
			}
		})
	}

}

// fileExists return true if a file exists.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

package s3

import (
	"context"
	"fmt"
	"io"

	"bitbucket.org/nevroz/dataflow"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Ensure that S3FetchService implements dataflow.Fetch.
var _ dataflow.Fetch = (*S3FetchService)(nil)

// S3FetchService represents a connection to the AWS S3 bucket.
type S3FetchService struct {
	S3Client *s3.Client
	Bucket   string
}

// NewS3FetchService initializes the S3FetchService struct.
func NewS3FetchService(bucket string) (*S3FetchService, error) {

	// AWS credentials loaded from environment.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	return &S3FetchService{
		S3Client: client,
		Bucket:   bucket,
	}, err
}

// Get method downloads the S3 object represented by key.
func (s *S3FetchService) Get(ctx context.Context, key string) ([]byte, error) {

	fmt.Printf("Downloading: %s \n", key)
	result, err := s.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	fmt.Printf("Finished downloading: %s \n", key)
	//Read the data out of object.
	return io.ReadAll(result.Body)

}

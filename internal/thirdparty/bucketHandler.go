package thirdparty

import (
	"fmt"
	"time"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
    // "github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws/endpoints"

	"BackendFramework/internal/config"
)


func newSession() (*session.Session) {
	defaultResolver := endpoints.DefaultResolver()
	s3CustResolverFn := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service == "s3" {
			return endpoints.ResolvedEndpoint{
				URL:           config.AWS_ENDPOINT,
			}, nil
		}

		return defaultResolver.EndpointFor(service, region, optFns...)
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(config.AWS_REGION),
			EndpointResolver: endpoints.ResolverFunc(s3CustResolverFn),
			Credentials: credentials.NewStaticCredentials(
				config.AWS_ACCESS_KEY_ID,
				config.AWS_SECRET_ACCESS_KEY,
				"",
			),
		},
	}))

	return sess
}

func GetFileBucket(fileLoc string) (string, error) {
	sess := newSession()
	s3Client := s3.New(sess)
	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(config.AWS_BUCKET_NAME),
		Key:    aws.String(fileLoc),
	})
	urlStr, err := req.Presign(24 * time.Hour)
	if err != nil {
		return "", err
	}

	return urlStr, nil
}

func UploadFileBucket(localFilePath, filename string) (string, error) {
	sess := newSession()

	// Create S3 service client
	s3Client := s3.New(sess)

	// Open the local file
	file, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open local file: %v", err)
	}
	defer file.Close()

	// Define the S3 object key (file name in S3)
	s3Key := filename

	// Upload the file to S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(config.AWS_BUCKET_NAME),
		Key:    aws.String(s3Key),
		Body:   file,
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	// Return the URL of the uploaded file in S3
	s3Url,err := GetFileBucket(s3Key)
	if err != nil {
		return "", fmt.Errorf("failed to get file from s3 : %v", err)
	}
	return s3Url, nil
}
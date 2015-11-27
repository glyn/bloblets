package blobstore

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	service    *s3.S3
	bucketName string
)

func init() {
	// Initialise S3 only when the server initialises.
	if !strings.HasSuffix(os.Args[0], "client") {
		// See: http://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
		service, bucketName = createService("s3.amazonaws.com", "us-east-1")
	}
}

func Present(key string) bool {
	return head(service, bucketName, key)
}

func Add(key, src string) {
	put(service, bucketName, key, src)
}

func Get(key, dest string) {
	get(service, bucketName, key, dest)
}

func Terminate() {
	defer service.DeleteBucket(&s3.DeleteBucketInput{Bucket: stringPtr(bucketName)})
	emptyBucket(service, bucketName)
}

func createService(endpoint, region string) (*s3.S3, string) {
	session := session.New()
	session.Config.Endpoint = stringPtr(endpoint)
	session.Config.Region = stringPtr(region)

	service := s3.New(session)
	bucketName := "com-github-glyn-bloblets-blobstore"

	_, err := service.HeadBucket(&s3.HeadBucketInput{Bucket: stringPtr(bucketName)})
	if err != nil {

		for i := 0; i < 10; i++ {
			_, err = service.CreateBucket(&s3.CreateBucketInput{Bucket: stringPtr(bucketName)})
			if err == nil {
				break
			}
		}
		if err != nil {
			panic(err)
		}

		for i := 0; i < 30; i++ {
			_, err = service.HeadBucket(&s3.HeadBucketInput{Bucket: stringPtr(bucketName)})
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if err != nil {
			panic(err)
		}
	}

	emptyBucket(service, bucketName)

	return service, bucketName
}

func put(service *s3.S3, bucketName string, key, path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = service.PutObject(&s3.PutObjectInput{
		Bucket: stringPtr(bucketName),
		Key:    stringPtr(key),
		Body:   f,
	})
	if err != nil {
		panic(err)
	}
}

func get(service *s3.S3, bucketName string, key, path string) {
	obj, err := service.GetObject(&s3.GetObjectInput{
		Bucket: stringPtr(bucketName),
		Key:    stringPtr(key),
	})
	if err != nil {
		panic(err)
	}
	defer obj.Body.Close()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = io.Copy(f, obj.Body)
	if err != nil {
		panic(err)
	}
}

func head(service *s3.S3, bucketName string, key string) bool {
	_, err := service.HeadObject(&s3.HeadObjectInput{
		Bucket: stringPtr(bucketName),
		Key:    stringPtr(key),
	})
	return err == nil
}

func emptyBucket(service *s3.S3, bucketName string) {
	objs, err := service.ListObjects(&s3.ListObjectsInput{Bucket: stringPtr(bucketName)})
	if err != nil {
		panic(err)
	}

	for _, o := range objs.Contents {
		_, err := service.DeleteObject(&s3.DeleteObjectInput{Bucket: stringPtr(bucketName), Key: o.Key})
		if err != nil {
			panic(err)
		}
	}
}

func stringPtr(s string) *string {
	return &s
}

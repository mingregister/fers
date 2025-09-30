package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

var _ Client = (*ossClient)(nil)

type ossClient struct {
	client     *oss.Client
	bucketName string
	workDir    string
}

// NewOSSClient creates a new OSS client using SDK v2
func NewOSSClient(endpoint, accessKeyID, accessKeySecret, bucketName, region, workDir string) (Client, error) {
	// Create credentials provider
	credentialsProvider := credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret)

	// Create OSS client configuration
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentialsProvider).
		WithRegion(region).
		WithEndpoint(endpoint)

	// Create OSS client
	client := oss.NewClient(cfg)

	workDir = NormalizePath(workDir)
	return &ossClient{
		client:     client,
		bucketName: bucketName,
		workDir:    workDir,
	}, nil
}

// List all object keys under given prefix
func (o *ossClient) List(ctx context.Context, prefix string) ([]string, error) {
	objects := make([]string, 0, 128)

	// Create list objects request
	request := &oss.ListObjectsV2Request{
		Bucket:  oss.Ptr(o.bucketName),
		Prefix:  oss.Ptr(o.getFullPath(prefix)),
		MaxKeys: int32(1000),
	}

	for {
		// List objects
		result, err := o.client.ListObjectsV2(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		// Collect object keys and remove workDir prefix
		for _, object := range result.Contents {
			if object.Key != nil {
				key := *object.Key
				// Remove workDir prefix from returned keys
				if o.workDir != "" && strings.HasPrefix(key, o.workDir+"/") {
					key = strings.TrimPrefix(key, o.workDir+"/")
				} else if o.workDir != "" && key == o.workDir {
					key = ""
				}
				objects = append(objects, key)
			}
		}

		// Check if there are more objects
		if !result.IsTruncated {
			break
		}

		// Set continuation token for next request
		if result.NextContinuationToken != nil {
			request.ContinuationToken = result.NextContinuationToken
		} else {
			break
		}
	}

	return objects, nil
}

// Upload object with given key and content
func (o *ossClient) Upload(ctx context.Context, key string, data []byte) error {
	reader := bytes.NewReader(data)

	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(o.bucketName),
		Key:    oss.Ptr(o.getFullPath(key)),
		Body:   reader,
	}

	_, err := o.client.PutObject(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to upload object %s: %w", key, err)
	}

	return nil
}

// Download object by key
func (o *ossClient) Download(ctx context.Context, key string) ([]byte, error) {
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(o.bucketName),
		Key:    oss.Ptr(o.getFullPath(key)),
	}

	result, err := o.client.GetObject(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to download object %s: %w", key, err)
	}
	defer result.Body.Close()

	// Read all data
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data %s: %w", key, err)
	}

	return data, nil
}

func (o *ossClient) getFullPath(key string) string {
	fullPath := fmt.Sprintf("%s/%s", o.workDir, key)
	return NormalizePath(fullPath)
}

func (o *ossClient) Delete(ctx context.Context, key string) error {
	return nil
}

// IsOSSKey checks if the given key looks like an OSS object key
func IsOSSKey(key string) bool {
	return strings.HasPrefix(key, "oss://") ||
		strings.Contains(key, ".oss-") ||
		(!strings.Contains(key, "\\") && !strings.HasPrefix(key, "/") && !strings.Contains(key, ":"))
}

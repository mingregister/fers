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

var _ Client = &ossClient{}

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

	workDir = strings.Replace(workDir, "//", "/", -1)
	workDir = strings.TrimPrefix(workDir, "/")
	return &ossClient{
		client:     client,
		bucketName: bucketName,
		workDir:    workDir,
	}, nil
}

// List all object keys under given prefix
func (o *ossClient) List(prefix string) ([]string, error) {
	var objects []string

	// Create list objects request
	request := &oss.ListObjectsV2Request{
		Bucket:  oss.Ptr(o.bucketName),
		Prefix:  oss.Ptr(o.getFullPath(prefix)),
		MaxKeys: int32(1000),
	}

	ctx := context.Background()

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
func (o *ossClient) Upload(key string, data []byte) error {
	reader := bytes.NewReader(data)

	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(o.bucketName),
		Key:    oss.Ptr(o.getFullPath(key)),
		Body:   reader,
	}

	ctx := context.Background()
	_, err := o.client.PutObject(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to upload object %s: %w", key, err)
	}

	return nil
}

// Download object by key
func (o *ossClient) Download(key string) ([]byte, error) {
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(o.bucketName),
		Key:    oss.Ptr(o.getFullPath(key)),
	}

	ctx := context.Background()
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
	// 如果 workDir 为空，直接返回 key
	if o.workDir == "" {
		return key
	}

	// 确保 workDir 不以 / 结尾，key 不以 / 开头
	workDir := strings.TrimSuffix(o.workDir, "/")
	cleanKey := strings.TrimPrefix(key, "/")

	// 如果 key 为空，只返回 workDir
	if cleanKey == "" {
		return workDir
	}

	// 组合路径
	fullPath := fmt.Sprintf("%s/%s", workDir, cleanKey)
	return strings.Replace(fullPath, "//", "/", -1)
}

func (o *ossClient) Delete(key string) error {
	return nil
}

// IsOSSKey checks if the given key looks like an OSS object key
func IsOSSKey(key string) bool {
	return strings.HasPrefix(key, "oss://") ||
		strings.Contains(key, ".oss-") ||
		(!strings.Contains(key, "\\") && !strings.HasPrefix(key, "/") && !strings.Contains(key, ":"))
}

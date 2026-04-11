// Package s3 wraps aws-sdk-go-v2/s3 with a small surface tailored to how
// XMart Cloud services use object storage: put bytes, delete by key,
// generate presigned GET URLs and compute public URLs.
//
// The wrapper is endpoint-aware so it can talk to MinIO or other
// S3-compatible servers by setting Config.Endpoint + Config.ForcePathStyle.
//
// Domain-specific logic (image compression, HMAC tokens, media tables)
// stays in each service; this package deliberately only handles the
// generic storage primitives.
package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Config describes how to connect to an S3-compatible backend.
type Config struct {
	// Region is the AWS region (e.g. "us-east-1"). Required.
	Region string
	// Bucket is the target bucket name. Required.
	Bucket string
	// AccessKey / SecretKey are the static credentials. If both are empty,
	// the SDK default credential chain is used (IAM role, env, ~/.aws/).
	AccessKey string
	SecretKey string

	// Endpoint overrides the service endpoint. Set this for MinIO or other
	// S3-compatible servers. Leave empty for AWS S3.
	Endpoint string
	// ForcePathStyle forces path-style addressing (required by MinIO).
	ForcePathStyle bool

	// Prefix is prepended to every key written by Put. Useful for sharing
	// a bucket between services. Leading/trailing slashes are trimmed.
	Prefix string

	// CDNBaseURL is the public URL prefix for objects (CloudFront, CDN).
	// When set, PublicURL uses it; otherwise PublicURL falls back to the
	// direct bucket URL.
	CDNBaseURL string

	// ACLPublic makes Put upload objects with public-read ACL. Default false.
	ACLPublic bool
}

// Client is the S3 wrapper. Safe for concurrent use.
type Client struct {
	cli        *awss3.Client
	bucket     string
	prefix     string
	cdnBaseURL string
	region     string
	aclPublic  bool
}

// New builds an S3 client from Config. It validates the required fields
// but does not issue any network calls at construction time — the first
// Put/Get will surface any credential/endpoint problems.
func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Region == "" {
		return nil, errors.New("xmart-platform/storage/s3: empty Region")
	}
	if cfg.Bucket == "" {
		return nil, errors.New("xmart-platform/storage/s3: empty Bucket")
	}

	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		loadOpts = append(loadOpts,
			awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
			),
		)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	s3Opts := func(o *awss3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.ForcePathStyle
	}

	return &Client{
		cli:        awss3.NewFromConfig(awsCfg, s3Opts),
		bucket:     cfg.Bucket,
		prefix:     strings.Trim(cfg.Prefix, "/"),
		cdnBaseURL: strings.TrimRight(cfg.CDNBaseURL, "/"),
		region:     cfg.Region,
		aclPublic:  cfg.ACLPublic,
	}, nil
}

// Raw returns the underlying aws-sdk-go-v2/s3 client for advanced usage
// (multipart uploads, bucket operations, custom request middlewares).
func (c *Client) Raw() *awss3.Client { return c.cli }

// Bucket returns the bucket name this client was built for.
func (c *Client) Bucket() string { return c.bucket }

// Put uploads a byte payload at the given key. The key is prefixed with
// Config.Prefix when set. The returned URL is the public URL computed by
// PublicURL for the final key.
func (c *Client) Put(ctx context.Context, key string, body []byte, contentType string) (string, error) {
	full := c.JoinKey(key)
	in := &awss3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(full),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	}
	if c.aclPublic {
		in.ACL = types.ObjectCannedACLPublicRead
		in.ContentDisposition = aws.String("inline")
	}
	if _, err := c.cli.PutObject(ctx, in); err != nil {
		return "", fmt.Errorf("s3 put: %w", err)
	}
	return c.PublicURL(full), nil
}

// Delete removes the object at the given (unprefixed) key.
func (c *Client) Delete(ctx context.Context, key string) error {
	full := c.JoinKey(key)
	_, err := c.cli.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(full),
	})
	if err != nil {
		return fmt.Errorf("s3 delete: %w", err)
	}
	return nil
}

// PresignGet returns a presigned GET URL for the given key. The URL is
// valid for `expires` — use this for private objects that need temporary
// access (downloads, media previews).
func (c *Client) PresignGet(ctx context.Context, key string, expires time.Duration) (string, error) {
	full := c.JoinKey(key)
	ps := awss3.NewPresignClient(c.cli)
	req, err := ps.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(full),
	}, awss3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("s3 presign: %w", err)
	}
	return req.URL, nil
}

// JoinKey applies Config.Prefix to a user-provided key. Exported so
// callers can compute full keys for indexing in their own metadata tables.
func (c *Client) JoinKey(key string) string {
	key = strings.TrimLeft(key, "/")
	if c.prefix == "" {
		return key
	}
	return c.prefix + "/" + key
}

// PublicURL returns the URL a browser would use to fetch the object. When
// CDNBaseURL is set it is used verbatim; otherwise the direct
// bucket.s3.region.amazonaws.com URL is returned.
func (c *Client) PublicURL(fullKey string) string {
	if c.cdnBaseURL != "" {
		return c.cdnBaseURL + "/" + fullKey
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.bucket, c.region, fullKey)
}

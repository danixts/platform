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

type Config struct {
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	Endpoint       string
	ForcePathStyle bool
	Prefix         string
	CDNBaseURL     string
	ACLPublic      bool
}

type Client struct {
	cli        *awss3.Client
	bucket     string
	prefix     string
	cdnBaseURL string
	region     string
	aclPublic  bool
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Region == "" {
		return nil, errors.New("platform/storage/s3: empty Region")
	}
	if cfg.Bucket == "" {
		return nil, errors.New("platform/storage/s3: empty Bucket")
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

func (c *Client) Raw() *awss3.Client { return c.cli }

func (c *Client) Bucket() string { return c.bucket }

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

func (c *Client) JoinKey(key string) string {
	key = strings.TrimLeft(key, "/")
	if c.prefix == "" {
		return key
	}
	return c.prefix + "/" + key
}

func (c *Client) PublicURL(fullKey string) string {
	if c.cdnBaseURL != "" {
		return c.cdnBaseURL + "/" + fullKey
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.bucket, c.region, fullKey)
}

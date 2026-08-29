package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

const (
	defaultPresignTTL = 10 * time.Minute
	defaultS3Region   = "us-east-1"
)

// s3Bucket is the AWS SDK v2 S3-compatible Bucket. One driver for MinIO, R2, S3.
// ops talks to Endpoint (docker DNS, R2 API). sign talks to the host the browser
// sees so SigV4 Host matches the PUT. No emulator branches.
type s3Bucket struct {
	name   string
	region string
	ops    *s3.Client
	sign   *s3.Client
	presign *s3.PresignClient
}

// NewS3Bucket builds a Bucket from Config. Does not probe the network; first
// Head/Get/Put is the health check.
func NewS3Bucket(ctx context.Context, cfg Config) (Bucket, error) {
	_ = ctx
	if err := validateS3Config(cfg); err != nil {
		return nil, err
	}
	region := cfg.Region
	if region == "" {
		region = defaultS3Region
	}

	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	awsCfg := aws.Config{
		Region:      region,
		Credentials: creds,
	}

	opsEndpoint, err := opsBaseEndpoint(cfg)
	if err != nil {
		return nil, err
	}
	signEndpoint, err := signerBaseEndpoint(cfg)
	if err != nil {
		return nil, err
	}

	ops := s3.NewFromConfig(awsCfg, s3ClientOptions(opsEndpoint, cfg.ForcePathStyle)...)
	var sign *s3.Client
	if signEndpoint == opsEndpoint {
		sign = ops
	} else {
		sign = s3.NewFromConfig(awsCfg, s3ClientOptions(signEndpoint, cfg.ForcePathStyle)...)
	}

	return &s3Bucket{
		name:    cfg.Bucket,
		region:  region,
		ops:     ops,
		sign:    sign,
		presign: s3.NewPresignClient(sign),
	}, nil
}

func validateS3Config(cfg Config) error {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return fmt.Errorf("storage: S3 bucket name is required")
	}
	if strings.TrimSpace(cfg.AccessKey) == "" || strings.TrimSpace(cfg.SecretKey) == "" {
		return fmt.Errorf("storage: S3 access key and secret key are required")
	}
	return nil
}

func s3ClientOptions(baseEndpoint string, forcePathStyle bool) []func(*s3.Options) {
	return []func(*s3.Options){
		func(o *s3.Options) {
			if baseEndpoint != "" {
				o.BaseEndpoint = aws.String(baseEndpoint)
			}
			o.UsePathStyle = forcePathStyle
			o.DisableLogOutputChecksumValidationSkipped = true
			// Default WhenSupported adds x-amz-checksum-* that browsers won't replay.
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
			o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
			if strings.HasPrefix(baseEndpoint, "http://") {
				o.EndpointOptions.DisableHTTPS = true
			}
		},
	}
}

func (b *s3Bucket) Put(ctx context.Context, key string, r io.Reader, opts PutOptions) error {
	if err := requireKey(key); err != nil {
		return err
	}
	input := &s3.PutObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
		Body:   r,
	}
	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}
	if opts.MaxBytes > 0 {
		input.Body = io.LimitReader(r, opts.MaxBytes+1)
	}
	_, err := b.ops.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("storage: put %q: %w", key, err)
	}
	return nil
}

func (b *s3Bucket) Get(ctx context.Context, key string) (io.ReadCloser, *ObjectInfo, error) {
	if err := requireKey(key); err != nil {
		return nil, nil, err
	}
	out, err := b.ops.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, nil, fmt.Errorf("storage: get %q: %w", key, ErrObjectNotFound)
		}
		return nil, nil, fmt.Errorf("storage: get %q: %w", key, err)
	}
	return out.Body, objectInfoFromGet(key, out), nil
}

func (b *s3Bucket) Head(ctx context.Context, key string) (*ObjectInfo, error) {
	if err := requireKey(key); err != nil {
		return nil, err
	}
	out, err := b.ops.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("storage: head %q: %w", key, ErrObjectNotFound)
		}
		return nil, fmt.Errorf("storage: head %q: %w", key, err)
	}
	return objectInfoFromHead(key, out), nil
}

func (b *s3Bucket) Delete(ctx context.Context, key string) error {
	if err := requireKey(key); err != nil {
		return err
	}
	_, err := b.ops.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("storage: delete %q: %w", key, err)
	}
	return nil
}

func (b *s3Bucket) PresignPut(ctx context.Context, key string, opts PresignPutOptions) (*PresignPut, error) {
	if err := requireKey(key); err != nil {
		return nil, err
	}
	ttl := presignTTL(opts.TTL)
	expires := time.Now().Add(ttl)

	input := &s3.PutObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	}
	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}

	out, err := b.presign.PresignPutObject(ctx, input, func(po *s3.PresignOptions) {
		po.Expires = ttl
		if opts.ContentType != "" {
			po.ClientOptions = append(po.ClientOptions, restoreContentType(opts.ContentType))
		}
	})
	if err != nil {
		return nil, fmt.Errorf("storage: presign put %q: %w", key, err)
	}

	headers := replayHeaders(out.SignedHeader, opts.ContentType)
	if opts.ContentType != "" && headers.Get("Content-Type") != opts.ContentType {
		return nil, fmt.Errorf("storage: presign put %q: Content-Type was not signed", key)
	}

	return &PresignPut{
		URL:     out.URL,
		Method:  http.MethodPut,
		Headers: headers,
		Expires: expires,
	}, nil
}

func (b *s3Bucket) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if err := requireKey(key); err != nil {
		return "", err
	}
	ttl = presignTTL(ttl)
	out, err := b.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	}, func(po *s3.PresignOptions) {
		po.Expires = ttl
	})
	if err != nil {
		return "", fmt.Errorf("storage: presign get %q: %w", key, err)
	}
	return out.URL, nil
}

// restoreContentType re-adds Content-Type after the SDK strips it on empty-body
// presign (RemoveContentTypeHeader when Content-Length is 0). Must run in
// Finalize before Signing so the header is in the SigV4 canonical request.
func restoreContentType(contentType string) func(*s3.Options) {
	return func(o *s3.Options) {
		o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
			m := &restoreContentTypeMiddleware{contentType: contentType}
			if err := stack.Finalize.Insert(m, "Signing", middleware.Before); err != nil {
				return stack.Finalize.Add(m, middleware.Before)
			}
			return nil
		})
	}
}

type restoreContentTypeMiddleware struct {
	contentType string
}

func (m *restoreContentTypeMiddleware) ID() string { return "RestoreContentTypeForPresign" }

func (m *restoreContentTypeMiddleware) HandleFinalize(
	ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler,
) (middleware.FinalizeOutput, middleware.Metadata, error) {
	if m.contentType != "" {
		if req, ok := in.Request.(*smithyhttp.Request); ok {
			req.Header.Set("Content-Type", m.contentType)
		}
	}
	return next.HandleFinalize(ctx, in)
}

// replayHeaders is what the browser must send. Host is implied by the URL.
func replayHeaders(signed http.Header, contentType string) http.Header {
	out := make(http.Header)
	for k, vs := range signed {
		if strings.EqualFold(k, "Host") || strings.EqualFold(k, "Authorization") {
			continue
		}
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	if contentType != "" && out.Get("Content-Type") == "" {
		out.Set("Content-Type", contentType)
	}
	return out
}

func presignTTL(d time.Duration) time.Duration {
	if d <= 0 {
		return defaultPresignTTL
	}
	return d
}

func requireKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if strings.HasPrefix(key, "/") || strings.Contains(key, "..") || strings.ContainsAny(key, "\\\x00") {
		return ErrInvalidKey
	}
	return nil
}

func objectInfoFromHead(key string, out *s3.HeadObjectOutput) *ObjectInfo {
	info := &ObjectInfo{Key: key}
	if out.ContentLength != nil {
		info.Size = *out.ContentLength
	}
	if out.ContentType != nil {
		info.ContentType = *out.ContentType
	}
	if out.ETag != nil {
		info.ETag = strings.Trim(*out.ETag, `"`)
	}
	return info
}

func objectInfoFromGet(key string, out *s3.GetObjectOutput) *ObjectInfo {
	info := &ObjectInfo{Key: key}
	if out.ContentLength != nil {
		info.Size = *out.ContentLength
	}
	if out.ContentType != nil {
		info.ContentType = *out.ContentType
	}
	if out.ETag != nil {
		info.ETag = strings.Trim(*out.ETag, `"`)
	}
	return info
}

func isNotFound(err error) bool {
	var nfe *types.NotFound
	if errors.As(err, &nfe) {
		return true
	}
	var nsk *types.NoSuchKey
	if errors.As(err, &nsk) {
		return true
	}
	var ae smithy.APIError
	if errors.As(err, &ae) {
		switch ae.ErrorCode() {
		case "NotFound", "NoSuchKey", "NoSuchBucket", "404":
			return true
		}
	}
	return false
}

// withScheme turns a host or URL into an absolute endpoint.
// "minio:9000" + UseTLS=false → "http://minio:9000"
// "http://minio:9000" is kept as-is.
func withScheme(raw string, useTLS bool) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, "://") {
		return strings.TrimRight(raw, "/")
	}
	scheme := "http"
	if useTLS {
		scheme = "https"
	}
	return scheme + "://" + strings.TrimRight(raw, "/")
}

// opsBaseEndpoint is S3_ENDPOINT. Empty means AWS default resolver.
func opsBaseEndpoint(cfg Config) (string, error) {
	return withScheme(cfg.Endpoint, cfg.UseTLS), nil
}

// signerBaseEndpoint is the host the browser will PUT to.
// PublicBaseURL "http://localhost:9000/thread-art" → "http://localhost:9000"
// (path is the bucket in path-style; putting it in BaseEndpoint would double it).
// Empty PublicBaseURL → same as ops.
func signerBaseEndpoint(cfg Config) (string, error) {
	if strings.TrimSpace(cfg.PublicBaseURL) == "" {
		return opsBaseEndpoint(cfg)
	}
	raw := strings.TrimSpace(cfg.PublicBaseURL)
	if !strings.Contains(raw, "://") {
		return withScheme(hostOnly(raw), cfg.UseTLS), nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("storage: PublicBaseURL: %w", err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("storage: PublicBaseURL missing host")
	}
	scheme := u.Scheme
	if scheme == "" {
		if cfg.UseTLS {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + u.Host, nil
}

func hostOnly(raw string) string {
	raw = strings.TrimSpace(raw)
	if i := strings.Index(raw, "/"); i >= 0 {
		raw = raw[:i]
	}
	return raw
}

// objectURL is the path-style or virtual-hosted URL for a key. Used by tests
// and as documentation of MinIO public URL shape: {base}/{bucket}/{key}.
func objectURL(baseEndpoint, bucket, key string, pathStyle bool) string {
	baseEndpoint = strings.TrimRight(baseEndpoint, "/")
	key = strings.TrimLeft(key, "/")
	if pathStyle {
		return baseEndpoint + "/" + bucket + "/" + key
	}
	u, err := url.Parse(baseEndpoint)
	if err != nil || u.Host == "" {
		return baseEndpoint + "/" + key
	}
	return u.Scheme + "://" + bucket + "." + u.Host + "/" + key
}

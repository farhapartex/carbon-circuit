package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"
	"github.com/minio/minio-go/v7/pkg/sse"
)

var ErrEncryptionUnavailable = errors.New(
	"object storage refused to enable server-side encryption, and evidence is never stored unencrypted")

type Settings struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseTLS    bool
	LinkTTL   time.Duration
}

type Store struct {
	client   *minio.Client
	bucket   string
	linkTTL  time.Duration
	encrypts encrypt.ServerSide
}

func Open(ctx context.Context, settings Settings) (*Store, error) {
	client, err := minio.New(settings.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(settings.AccessKey, settings.SecretKey, ""),
		Secure: settings.UseTLS,
	})
	if err != nil {
		return nil, fmt.Errorf("open object storage: %w", err)
	}

	store := &Store{
		client:   client,
		bucket:   settings.Bucket,
		linkTTL:  settings.LinkTTL,
		encrypts: encrypt.NewSSE(),
	}

	if err := store.prepare(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) prepare(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %s: %w", s.bucket, err)
	}

	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket %s: %w", s.bucket, err)
		}
	}

	if err := s.client.SetBucketEncryption(ctx, s.bucket, sse.NewConfigurationSSES3()); err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptionUnavailable, err)
	}

	configuration, err := s.client.GetBucketEncryption(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptionUnavailable, err)
	}
	if len(configuration.Rules) == 0 {
		return ErrEncryptionUnavailable
	}

	return nil
}

func (s *Store) Put(ctx context.Context, key string, content []byte, mediaType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(content), int64(len(content)),
		minio.PutObjectOptions{
			ContentType:          mediaType,
			ServerSideEncryption: s.encrypts,
		})
	if err != nil {
		return fmt.Errorf("store object %s: %w", key, err)
	}

	return nil
}

func (s *Store) Remove(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object %s: %w", key, err)
	}

	return nil
}

func (s *Store) SignedLink(
	ctx context.Context,
	key, fileName, mediaType string,
) (string, time.Time, error) {
	parameters := url.Values{}
	parameters.Set("response-content-disposition",
		fmt.Sprintf("attachment; filename=%q", fileName))
	parameters.Set("response-content-type", mediaType)

	signed, err := s.client.PresignedGetObject(ctx, s.bucket, key, s.linkTTL, parameters)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign download link for %s: %w", key, err)
	}

	return signed.String(), time.Now().UTC().Add(s.linkTTL), nil
}

func (s *Store) Reachable(ctx context.Context) bool {
	_, err := s.client.BucketExists(ctx, s.bucket)
	return err == nil
}

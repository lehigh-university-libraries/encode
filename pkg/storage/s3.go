package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Enabled      bool   `yaml:"enabled"`
	Bucket       string `yaml:"bucket"`
	Region       string `yaml:"region"`
	Prefix       string `yaml:"prefix"`
	ManifestPath string `yaml:"manifest_path"`
}

type S3Uploader struct {
	client *s3.Client
	config S3Config
}

type QuickSightManifest struct {
	FileLocations        []FileLocation       `json:"fileLocations"`
	GlobalUploadSettings GlobalUploadSettings `json:"globalUploadSettings"`
}

type FileLocation struct {
	URIs []string `json:"URIs,omitempty"`
}

type GlobalUploadSettings struct {
	Format         string `json:"format"`
	Delimiter      string `json:"delimiter,omitempty"`
	TextQualifier  string `json:"textqualifier,omitempty"`
	ContainsHeader string `json:"containsHeader"`
}

func NewS3Uploader(s3Config S3Config) (*S3Uploader, error) {
	if !s3Config.Enabled {
		return &S3Uploader{config: s3Config}, nil
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(s3Config.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Uploader{
		client: client,
		config: s3Config,
	}, nil
}

// UploadFile uploads a local file to S3
func (u *S3Uploader) UploadFile(localPath, reportName string) (string, error) {
	if !u.config.Enabled {
		slog.Debug("S3 upload disabled, skipping")
		return "", nil
	}

	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", localPath, err)
	}
	defer file.Close()

	// Generate S3 key: prefix/report_name/filename
	filename := filepath.Base(localPath)
	key := filepath.Join(u.config.Prefix, reportName, filename)
	// Normalize path separators for S3 (always use forward slash)
	key = strings.ReplaceAll(key, "\\", "/")

	slog.Info("Uploading file to S3", "localPath", localPath, "bucket", u.config.Bucket, "key", key)

	_, err = u.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(u.config.Bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return the S3 URI
	uri := fmt.Sprintf("s3://%s/%s", u.config.Bucket, key)
	slog.Info("Successfully uploaded to S3", "uri", uri)
	return uri, nil
}

// GenerateManifest creates or updates a QuickSight-compatible manifest file
// It appends new URIs to existing ones to maintain historical data for QuickSight
func (u *S3Uploader) GenerateManifest(reportName string, newS3URI string) error {
	if !u.config.Enabled || u.config.ManifestPath == "" {
		return nil
	}

	// Create manifest directory if it doesn't exist
	manifestDir := filepath.Join(u.config.ManifestPath, reportName)
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	// Use a static manifest filename per report (not date-based)
	manifestFile := filepath.Join(manifestDir, "manifest.json")

	// Load existing manifest if it exists
	var existingURIs []string
	if existingData, err := os.ReadFile(manifestFile); err == nil {
		var existingManifest QuickSightManifest
		if err := json.Unmarshal(existingData, &existingManifest); err == nil {
			if len(existingManifest.FileLocations) > 0 {
				existingURIs = existingManifest.FileLocations[0].URIs
			}
		}
	}

	// Append new URI if not already present
	allURIs := existingURIs
	isDuplicate := false
	for _, uri := range existingURIs {
		if uri == newS3URI {
			isDuplicate = true
			break
		}
	}
	if !isDuplicate {
		allURIs = append(allURIs, newS3URI)
	}

	// Create updated manifest
	manifest := QuickSightManifest{
		FileLocations: []FileLocation{
			{
				URIs: allURIs,
			},
		},
		GlobalUploadSettings: GlobalUploadSettings{
			Format:         "CSV",
			Delimiter:      ",",
			TextQualifier:  "\"",
			ContainsHeader: "true",
		},
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestFile, manifestJSON, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	slog.Info("Generated QuickSight manifest", "file", manifestFile, "totalURIs", len(allURIs))
	return nil
}

// UploadManifestForReport uploads the manifest file for a specific report to S3
func (u *S3Uploader) UploadManifestForReport(reportName string) (string, error) {
	if !u.config.Enabled {
		return "", nil
	}

	manifestDir := filepath.Join(u.config.ManifestPath, reportName)
	localManifestPath := filepath.Join(manifestDir, "manifest.json")

	// Check if manifest exists
	if _, err := os.Stat(localManifestPath); os.IsNotExist(err) {
		return "", fmt.Errorf("manifest file does not exist: %s", localManifestPath)
	}

	file, err := os.Open(localManifestPath)
	if err != nil {
		return "", fmt.Errorf("failed to open manifest file %s: %w", localManifestPath, err)
	}
	defer file.Close()

	key := filepath.Join(u.config.Prefix, "manifests", reportName, "manifest.json")
	key = strings.ReplaceAll(key, "\\", "/")

	slog.Info("Uploading manifest to S3", "localPath", localManifestPath, "bucket", u.config.Bucket, "key", key)

	_, err = u.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(u.config.Bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload manifest to S3: %w", err)
	}

	manifestURL := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/%s", u.config.Bucket, u.config.Region, key)
	slog.Info("Successfully uploaded manifest to S3", "url", manifestURL)
	return manifestURL, nil
}

package ptah_agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/docker/docker/api/types/mount"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/busybox"
	"github.com/ptah-sh/ptah-agent/internal/app/ptah-agent/docker/config"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (e *taskExecutor) createS3Storage(ctx context.Context, req *t.CreateS3StorageReq) (*t.CreateS3StorageRes, error) {
	var res t.CreateS3StorageRes

	decryptedSecretKey, err := e.decryptValue(ctx, req.S3StorageSpec.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("create s3 storage: decrypt secret key: %w", err)
	}

	decryptedSpec := req.S3StorageSpec
	decryptedSpec.SecretKey = decryptedSecretKey

	data, err := json.Marshal(decryptedSpec)
	if err != nil {
		return nil, fmt.Errorf("create s3 storage: marshal spec: %w", err)
	}

	req.SwarmConfigSpec.Data = data

	config, err := e.docker.ConfigCreate(ctx, req.SwarmConfigSpec)
	if err != nil {
		return nil, fmt.Errorf("create s3 storage: create config: %w", err)
	}

	res.Docker.ID = config.ID

	return &res, nil
}

func (e *taskExecutor) checkS3Storage(ctx context.Context, req *t.CheckS3StorageReq) (*t.CheckS3StorageRes, error) {
	envVars := []string{
		fmt.Sprintf("ARCHIVE_FORMAT=%s", "tar.gz"),
	}

	_, err := e.uploadS3FileWithHelper(ctx, []mount.Mount{}, envVars, req.S3StorageConfigName, "/tmp/check-access.txt", ".check-access.tar.gz")
	if err != nil {
		return nil, err
	}

	return &t.CheckS3StorageRes{}, nil
}

func (e *taskExecutor) s3upload(ctx context.Context, req *t.S3UploadReq) (*t.S3UploadRes, error) {
	envVars := []string{
		fmt.Sprintf("ARCHIVE_FORMAT=%s", req.Archive.Format),
	}

	output, err := e.uploadS3FileWithHelper(ctx, []mount.Mount{req.VolumeSpec}, envVars, req.S3StorageConfigName, req.SrcFilePath, req.DestFilePath)
	if err != nil {
		return nil, err
	}

	return &t.S3UploadRes{
		Output: output,
	}, nil
}

func (e *taskExecutor) s3download(ctx context.Context, req *t.S3DownloadReq) (*t.S3DownloadRes, error) {
	envVars := []string{}

	logs, err := e.runS3Cmd(ctx, []mount.Mount{req.VolumeSpec}, envVars, "/ptah/bin/s3_download.sh", req.S3StorageConfigName, req.SrcFilePath, req.DestFilePath)
	if err != nil {
		return nil, fmt.Errorf("download from s3: %w", err)
	}

	return &t.S3DownloadRes{
		Output: logs,
	}, nil
}

func (e *taskExecutor) uploadS3FileWithHelper(ctx context.Context, mounts []mount.Mount, envVars []string, s3StorageConfigName, srcFilePath, destFilePath string) ([]string, error) {
	return e.runS3Cmd(ctx, mounts, envVars, "/ptah/bin/s3_upload.sh", s3StorageConfigName, srcFilePath, destFilePath)
}

func (e *taskExecutor) runS3Cmd(ctx context.Context, mounts []mount.Mount, envVars []string, s3Script, s3StorageConfigName, srcFilePath, destFilePath string) ([]string, error) {
	credentialsConfig, err := config.GetByName(ctx, e.docker, s3StorageConfigName)
	if err != nil {
		return nil, fmt.Errorf("check s3 storage: get config: %w", err)
	}

	var s3StorageSpec t.S3StorageSpec
	err = json.Unmarshal(credentialsConfig.Spec.Data, &s3StorageSpec)
	if err != nil {
		return nil, fmt.Errorf("check s3 storage: unmarshal config: %w", err)
	}

	if strings.Index(destFilePath, s3StorageSpec.PathPrefix) == 0 {
		destFilePath = destFilePath[len(s3StorageSpec.PathPrefix):]
	}

	vars := []string{
		fmt.Sprintf("S3_ACCESS_KEY=%s", s3StorageSpec.AccessKey),
		fmt.Sprintf("S3_SECRET_KEY=%s", s3StorageSpec.SecretKey),
		fmt.Sprintf("S3_ENDPOINT=%s", s3StorageSpec.Endpoint),
		fmt.Sprintf("S3_REGION=%s", s3StorageSpec.Region),
		fmt.Sprintf("S3_BUCKET=%s", s3StorageSpec.Bucket),
		fmt.Sprintf("PATH_PREFIX=%s", strings.Trim(s3StorageSpec.PathPrefix, "/")),
		fmt.Sprintf("SRC_FILE_PATH=%s", strings.TrimPrefix(srcFilePath, "/")),
		fmt.Sprintf("DEST_FILE_PATH=%s", strings.TrimPrefix(destFilePath, "/")),
	}

	vars = append(vars, envVars...)

	box := busybox.New(e.docker)

	if err := box.Start(ctx, &busybox.Config{
		Cmd:     s3Script,
		EnvVars: vars,
		Mounts:  mounts,
	}); err != nil {
		return nil, fmt.Errorf("check s3 storage: start busybox: %w", err)
	}
	defer box.Stop(ctx)

	if err := box.Wait(ctx); err != nil {
		return nil, fmt.Errorf("check s3 storage: wait busybox: %w", err)
	}

	logs, err := box.Logs(ctx)
	if err != nil {
		return nil, fmt.Errorf("check s3 storage: get logs: %w", err)
	}

	return logs, nil
}

func (e *taskExecutor) removeS3Files(ctx context.Context, req *t.S3RemoveReq) (*t.S3RemoveRes, error) {
	log := Logger(ctx)

	credentialsConfig, err := config.GetByName(ctx, e.docker, req.S3StorageConfigName)
	if err != nil {
		return nil, fmt.Errorf("remove s3 file: get config: %w", err)
	}

	var s3StorageSpec t.S3StorageSpec
	err = json.Unmarshal(credentialsConfig.Spec.Data, &s3StorageSpec)
	if err != nil {
		return nil, fmt.Errorf("remove s3 file: unmarshal config: %w", err)
	}

	s3client := s3.NewFromConfig(aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(
			s3StorageSpec.AccessKey,
			s3StorageSpec.SecretKey,
			"",
		),
		Region:       s3StorageSpec.Region,
		BaseEndpoint: aws.String("https://" + s3StorageSpec.Endpoint),
	})

	fullPath := strings.Trim(s3StorageSpec.PathPrefix, "/") + "/" + strings.TrimPrefix(req.FilePath, "/")

	log.Debug("remove s3 file", "bucket", s3StorageSpec.Bucket, "full_path", fullPath)

	_, err = s3client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3StorageSpec.Bucket),
		Key:    aws.String(fullPath),
	})

	if err != nil {
		return nil, fmt.Errorf("remove s3 file: delete object: %w", err)
	}

	return &t.S3RemoveRes{}, nil
}

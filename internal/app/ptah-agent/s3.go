package ptah_agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
	"io"
	"net/url"
	path2 "path"
	"strings"
)

func (e *taskExecutor) createS3Storage(ctx context.Context, req *t.CreateS3StorageReq) (*t.CreateS3StorageRes, error) {
	var res t.CreateS3StorageRes

	if req.S3StorageSpec.AccessKey == "" || req.S3StorageSpec.SecretKey == "" {
		if req.PrevConfigName == "" {
			return nil, fmt.Errorf("create s3 storage: prev config name is empty - empty credentials")
		}

		prev, err := e.getConfigByName(ctx, req.PrevConfigName)
		if err != nil {
			return nil, err
		}

		var prevSpec t.S3StorageSpec
		err = json.Unmarshal(prev.Spec.Data, &prevSpec)
		if err != nil {
			return nil, fmt.Errorf("create s3 storage: unmarshal prev config: %w", err)
		}

		req.S3StorageSpec.AccessKey = prevSpec.AccessKey
		req.S3StorageSpec.SecretKey = prevSpec.SecretKey
	}

	data, err := json.Marshal(req.S3StorageSpec)
	if err != nil {
		return nil, err
	}

	req.SwarmConfigSpec.Data = data

	config, err := e.docker.ConfigCreate(ctx, req.SwarmConfigSpec)
	if err != nil {
		return nil, err
	}

	res.Docker.ID = config.ID

	return &res, nil
}

func (e *taskExecutor) checkS3Storage(ctx context.Context, req *t.CheckS3StorageReq) (*t.CheckS3StorageRes, error) {
	err := e.uploadToS3(ctx, ".check-access", bytes.NewReader([]byte("https://ptah.sh")), req.S3StorageConfigName)
	if err != nil {
		return nil, err
	}

	return &t.CheckS3StorageRes{}, nil
}

type staticResolver struct {
	URL string
}

func (r *staticResolver) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (smithyendpoints.Endpoint, error) {
	uri, err := url.Parse(r.URL)
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}

	return smithyendpoints.Endpoint{
		URI: *uri,
	}, nil
}

func (e *taskExecutor) uploadToS3(ctx context.Context, path string, body io.Reader, s3ConfigName string) error {
	credentialsConfig, err := e.getConfigByName(ctx, s3ConfigName)
	if err != nil {
		return err
	}

	var s3StorageSpec t.S3StorageSpec
	err = json.Unmarshal(credentialsConfig.Spec.Data, &s3StorageSpec)
	if err != nil {
		return err
	}

	credentialsProvider := credentials.NewStaticCredentialsProvider(s3StorageSpec.AccessKey, s3StorageSpec.SecretKey, "")
	config, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(s3StorageSpec.Region), awsConfig.WithCredentialsProvider(credentialsProvider))
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(config, s3.WithEndpointResolverV2(&staticResolver{
		URL: s3StorageSpec.Endpoint,
	}))

	//expires := time.Now().Add(time.Duration(30) * time.Second).UTC()
	filePath := strings.Trim(path2.Join(s3StorageSpec.PathPrefix, path), "/")

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s3StorageSpec.Bucket,
		Key:    &filePath,
		Body:   body,
		//Expires: &expires,
	})

	return err
}

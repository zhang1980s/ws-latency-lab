package resources

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecr"
	
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateEcrRepository creates an ECR repository for the WebSocket server Docker image
func CreateEcrRepository(ctx *pulumi.Context, cfg *config.Config) (*ecr.Repository, error) {
	// Get ECR configuration
	ecrConfig := config.GetEcrConfig()
	
	// Create ECR repository
	repo, err := ecr.NewRepository(ctx, "ws-server-repo", &ecr.RepositoryArgs{
		Name: pulumi.String(ecrConfig.Name),
		ImageScanningConfiguration: &ecr.RepositoryImageScanningConfigurationArgs{
			ScanOnPush: pulumi.Bool(true),
		},
		ImageTagMutability: pulumi.String("MUTABLE"),
		Tags:               utils.ApplyTags(ctx, "ws-server-repo", utils.GetNamedTags("ws-server-repo", cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}
	
	// Add lifecycle policy
	_, err = ecr.NewLifecyclePolicy(ctx, "ws-server-lifecycle", &ecr.LifecyclePolicyArgs{
		Repository: repo.Name,
		Policy: pulumi.String(`{
			"rules": [
				{
					"rulePriority": 1,
					"description": "Keep last 10 images",
					"selection": {
						"tagStatus": "any",
						"countType": "imageCountMoreThan",
						"countNumber": 10
					},
					"action": {
						"type": "expire"
					}
				}
			]
		}`),
	})
	if err != nil {
		return nil, err
	}
	
	return repo, nil
}

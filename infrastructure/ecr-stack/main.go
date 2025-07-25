package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/ecr-stack/resources"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Load configuration
		cfg, err := config.LoadConfig(ctx)
		if err != nil {
			return err
		}

		// Create ECR repository
		repo, err := resources.CreateEcrRepository(ctx, cfg)
		if err != nil {
			return err
		}

		// Export outputs
		ctx.Export("repositoryUrl", repo.RepositoryUrl)
		ctx.Export("repositoryName", repo.Name)

		return nil
	})
}

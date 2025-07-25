package resources

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateEcsResources creates ECS resources for client containers
func CreateEcsResources(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc, subnets []pulumi.IDOutput, sg *ec2.SecurityGroup, endpoint *ec2.VpcEndpoint) (*ecs.Cluster, *ecs.TaskDefinition, *ecs.Service, error) {
	// Get ECR repository reference
	ecrStackRef, err := pulumi.NewStackReference(ctx, "zhang1980s/ws-latency-ecr/dev", nil)
	if err != nil {
		return nil, nil, nil, err
	}

	ecrRepoUrl := ecrStackRef.GetOutput(pulumi.String("repositoryUrl"))

	// Create ECS cluster
	clusterName := config.FormatResourceName(cfg.Project, "client-cluster")
	cluster, err := ecs.NewCluster(ctx, clusterName, &ecs.ClusterArgs{
		Name: pulumi.String(clusterName),
		Settings: ecs.ClusterSettingArray{
			&ecs.ClusterSettingArgs{
				Name:  pulumi.String("containerInsights"),
				Value: pulumi.String("enabled"),
			},
		},
		Tags: utils.ApplyTags(ctx, clusterName, utils.GetNamedTags(clusterName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Create IAM role for ECS task execution
	executionRoleName := config.FormatResourceName(clusterName, "execution-role")
	executionRole, err := iam.NewRole(ctx, executionRoleName, &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Action": "sts:AssumeRole",
				"Principal": {
					"Service": "ecs-tasks.amazonaws.com"
				},
				"Effect": "Allow",
				"Sid": ""
			}]
		}`),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
		},
		Tags: utils.ApplyTags(ctx, executionRoleName, utils.GetNamedTags(executionRoleName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Create IAM role for ECS task
	taskRoleName := config.FormatResourceName(clusterName, "task-role")
	taskRole, err := iam.NewRole(ctx, taskRoleName, &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Action": "sts:AssumeRole",
				"Principal": {
					"Service": "ecs-tasks.amazonaws.com"
				},
				"Effect": "Allow",
				"Sid": ""
			}]
		}`),
		InlinePolicies: iam.RoleInlinePolicyArray{
			&iam.RoleInlinePolicyArgs{
				Name: pulumi.String("client-policy"),
				Policy: pulumi.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Action": [
							"logs:CreateLogGroup",
							"logs:CreateLogStream",
							"logs:PutLogEvents",
							"logs:DescribeLogStreams"
						],
						"Resource": "arn:aws:logs:*:*:*"
					}]
				}`),
			},
		},
		Tags: utils.ApplyTags(ctx, taskRoleName, utils.GetNamedTags(taskRoleName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Create ECS task definition
	taskDefName := config.FormatResourceName(cfg.Project, "client-task")

	// Use the endpoint DNS entries to configure the WebSocket server URL
	endpointDns := endpoint.DnsEntries.Index(pulumi.Int(0)).DnsName().ApplyT(func(dnsName string) string {
		return dnsName
	}).(pulumi.StringOutput)

	// Create container definition with environment variables
	containerDefinition := pulumi.All(ecrRepoUrl, endpointDns).ApplyT(
		func(args []interface{}) string {
			repoUrl := args[0].(string)
			dnsName := args[1].(string)

			return fmt.Sprintf(`[
				{
					"name": "ws-latency-client",
					"image": "%s:latest",
					"essential": true,
					"portMappings": [
						{
							"containerPort": %d,
							"hostPort": %d,
							"protocol": "tcp"
						},
						{
							"containerPort": 9090,
							"hostPort": 9090,
							"protocol": "tcp"
						}
					],
					"environment": [
						{
							"name": "WS_SERVER_URL",
							"value": "wss://%s:%d"
						},
						{
							"name": "CLIENT_MODE",
							"value": "true"
						},
						{
							"name": "METRICS_ENABLED",
							"value": "true"
						}
					],
					"logConfiguration": {
						"logDriver": "awslogs",
						"options": {
							"awslogs-group": "/ecs/ws-latency-client",
							"awslogs-region": "%s",
							"awslogs-stream-prefix": "ecs",
							"awslogs-create-group": "true"
						}
					}
				}
			]`, repoUrl, config.WebSocketPort, config.WebSocketPort, dnsName, config.NlbPort, cfg.Region)
		},
	).(pulumi.StringOutput)

	taskDef, err := ecs.NewTaskDefinition(ctx, taskDefName, &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String(taskDefName),
		Cpu:                     pulumi.String("256"),
		Memory:                  pulumi.String("512"),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
		ExecutionRoleArn:        executionRole.Arn,
		TaskRoleArn:             taskRole.Arn,
		ContainerDefinitions:    containerDefinition,
		Tags:                    utils.ApplyTags(ctx, taskDefName, utils.GetNamedTags(taskDefName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Create ECS service
	serviceName := config.FormatResourceName(cfg.Project, "client-service")

	// Convert subnets from []pulumi.IDOutput to pulumi.StringArray
	subnetIDs := pulumi.StringArray{}
	for _, subnet := range subnets {
		subnetIDs = append(subnetIDs, subnet.ToStringOutput())
	}

	service, err := ecs.NewService(ctx, serviceName, &ecs.ServiceArgs{
		Cluster:        cluster.Arn,
		DesiredCount:   pulumi.Int(2), // Run 2 instances for testing
		LaunchType:     pulumi.String("FARGATE"),
		TaskDefinition: taskDef.Arn,
		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			AssignPublicIp: pulumi.Bool(true),
			Subnets:        subnetIDs,
			SecurityGroups: pulumi.StringArray{sg.ID()},
		},
		Tags: utils.ApplyTags(ctx, serviceName, utils.GetNamedTags(serviceName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return cluster, taskDef, service, nil
}

package resources

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/cloudquery/plugin-sdk/schema"
	"github.com/cloudquery/plugin-sdk/transformers"
	"github.com/guardian/cq-source-cloudformation-template-summary/client"
)

type TemplateSummary struct {
	Metadata  *string
	StackName *string
	StackId   *string
}

func TemplateSummaries() *schema.Table {
	tableName := "cloudformation_template_summaries"
	return &schema.Table{
		Name:      tableName,
		Resolver:  fetchTemplateSummaries,
		Transform: transformers.TransformWithStruct(&TemplateSummary{}),
		Multiplex: client.ServiceAccountRegionMultiplexer(tableName, "cloudformation"),
	}
}

// fetchTemplateSummaries fetches a list of template summaries from the CloudFormation API
func fetchTemplateSummaries(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- interface{}) error {
	log.Println("fetching template summaries...")

	c := meta.(*client.Client)
	cfnClient := c.CloudformationClient()

	stacks, err := cfnClient.ListStacks(ctx, &cloudformation.ListStacksInput{})
	if err != nil {
		return err
	}

	for _, stack := range stacks.StackSummaries {
		stackName := stack.StackName

		input := &cloudformation.GetTemplateSummaryInput{StackName: stackName}
		summary, err := cfnClient.GetTemplateSummary(ctx, input)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				continue
			}

			log.Printf("unable to get template summary for stack %s: %v", *stackName, err)
			continue
		}

		table := &TemplateSummary{
			Metadata:  summary.Metadata,
			StackName: stack.StackName,
			StackId:   stack.StackId,
		}
		res <- table
	}
	return nil
}

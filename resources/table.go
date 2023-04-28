package resources

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/cloudquery/plugin-sdk/schema"
	"github.com/cloudquery/plugin-sdk/transformers"
	"github.com/guardian/cq-source-cloudformation-template-summary/client"
)

func SampleTable() *schema.Table {
	return &schema.Table{
		Name:      "cloudformation_template_summary_sample_table",
		Resolver:  fetchTemplateSummaries,
		Transform: transformers.TransformWithStruct(&cloudformation.GetTemplateSummaryOutput{}),
	}
}

// returns template summaries for all stacks in the account
func fetchTemplateSummaries(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- interface{}) error {
	c := meta.(*client.Client)
	stacks, err := c.CfnClient.ListStacks(ctx, &cloudformation.ListStacksInput{})

	if err != nil {
		return err
	}

	for _, stack := range stacks.StackSummaries {
		stackName := stack.StackName
		input := &cloudformation.GetTemplateSummaryInput{StackName: stackName}
		summary, err := c.CfnClient.GetTemplateSummary(ctx, input)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				continue
			}
			return err
		}
		res <- summary
	}
	return nil
}

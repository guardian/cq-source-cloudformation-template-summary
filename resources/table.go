package resources

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/cloudquery/plugin-sdk/schema"
	"github.com/cloudquery/plugin-sdk/transformers"
	"github.com/guardian/cq-source-cloudformation-template-summary/client"
)

type TemplateSummary struct {
	Metadata  string
	StackName string
	StackId   string
}

func SampleTable() *schema.Table {
	return &schema.Table{
		Name:      "cloudformation_template_summary_sample_table",
		Resolver:  fetchTemplateSummaries,
		Transform: transformers.TransformWithStruct(&TemplateSummary{}),
	}
}

// fetchTemplateSummaries fetches a list of template summaries from the CloudFormation API
func fetchTemplateSummaries(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- interface{}) error {
	// Create a new client using the meta object passed in
	c := meta.(*client.Client)

	// Call the ListStacks API to get a list of stacks
	stacks, err := c.CfnClient.ListStacks(ctx, &cloudformation.ListStacksInput{})
	if err != nil {
		return err
	}

	// Iterate through each stack in the list of stacks
	for _, stack := range stacks.StackSummaries {
		// Call the GetTemplateSummary API to get the template summary for the stack
		stackName := stack.StackName
		input := &cloudformation.GetTemplateSummaryInput{StackName: stackName}
		summary, err := c.CfnClient.GetTemplateSummary(ctx, input)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				continue
			}
			return err
		}

		// Create a new TemplateSummary object and populate it with the information from the API call
		table := &TemplateSummary{
			Metadata:  *summary.Metadata,
			StackName: *stack.StackName,
			StackId:   *stack.StackId,
		}
		res <- table
	}
	return nil
}

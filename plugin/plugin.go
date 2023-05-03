package plugin

import (
	"github.com/cloudquery/plugin-sdk/plugins/source"
	"github.com/cloudquery/plugin-sdk/schema"
	"github.com/guardian/cq-source-cloudformation-template-summary/client"
	"github.com/guardian/cq-source-cloudformation-template-summary/resources"
)

var (
	Version = "development"
)

func Plugin() *source.Plugin {
	return source.NewPlugin(
		"guardian-cloudformation-template-summary",
		Version,
		schema.Tables{
			resources.TemplateSummaries(),
		},
		client.New,
	)
}

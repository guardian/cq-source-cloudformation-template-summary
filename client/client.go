package client

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/cloudquery/plugin-sdk/plugins/source"
	"github.com/cloudquery/plugin-sdk/schema"
	"github.com/cloudquery/plugin-sdk/specs"
	"github.com/rs/zerolog"
)

type ClientsAccountRegionMap map[string]map[string]*cloudformation.Client
type Client struct {
	Logger                  zerolog.Logger
	ClientsAccountRegionMap ClientsAccountRegionMap
	Account                 string
	Region                  string
}

func (c Client) CloudformationClient() *cloudformation.Client {
	return c.ClientsAccountRegionMap[c.Account][c.Region]
}

func (c *Client) ID() string {
	// TODO: Change to either your plugin name or a unique dynamic identifier
	return "ID"
}

func New(ctx context.Context, logger zerolog.Logger, s specs.Source, opts source.Options) (schema.ClientMeta, error) {
	log.Println("creating the client!")

	var pluginSpec Spec

	if err := s.UnmarshalSpec(&pluginSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin spec: %w", err)
	}

	if len(pluginSpec.Accounts) > 0 {
		clients, err := clientsForAccounts(ctx, pluginSpec.Accounts, pluginSpec.Regions)

		if err != nil {
			return nil, fmt.Errorf("failed to create clients for account %v: %w", pluginSpec.Accounts, err)
		}

		return &Client{
			Logger:                  logger,
			ClientsAccountRegionMap: clients,
		}, nil
	}

	// else in actual AWS!

	clients, err := clientsForOrganisationUnits(ctx, pluginSpec.Organization, pluginSpec.Regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create clients for org %v: %w", pluginSpec.Organization, err)
	}

	return &Client{
		Logger:                  logger,
		ClientsAccountRegionMap: clients,
	}, nil
}

func clientsForAccounts(ctx context.Context, accounts []Account, regions []string) (ClientsAccountRegionMap, error) {
	clients := ClientsAccountRegionMap{}

	for _, account := range accounts {
		cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(account.LocalProfile))
		if err != nil {
			return nil, err
		}

		if clients[account.ID] == nil {
			clients[account.ID] = map[string]*cloudformation.Client{}
		}

		for _, region := range regions {
			clients[account.ID][region] = cloudformation.NewFromConfig(cfg, func(o *cloudformation.Options) { o.Region = region })
		}
	}

	return clients, nil
}

func clientsForOrganisationUnits(ctx context.Context, org *AwsOrg, regions []string) (ClientsAccountRegionMap, error) {
	clients := ClientsAccountRegionMap{}

	topLevelConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return nil, err
	}

	orgClient := organizations.NewFromConfig(topLevelConfig)

	accounts, err := getOUAccounts(ctx, orgClient, org)
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		roleARN := arn.ARN{
			Partition: "aws",
			Service:   "iam",
			Region:    "",
			AccountID: *account.Id,
			Resource:  "role/" + org.ChildAccountRoleName,
		}.String()

		//assume a role in each account
		creds, err := sts.NewFromConfig(topLevelConfig).AssumeRole(ctx, &sts.AssumeRoleInput{
			RoleArn:         aws.String(roleARN),
			RoleSessionName: aws.String("cloudquery"),
		})
		if err != nil {
			return nil, err
		}

		cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(*creds.Credentials.AccessKeyId, *creds.Credentials.SecretAccessKey, *creds.Credentials.SessionToken)))
		if err != nil {
			return nil, err
		}

		if clients[*account.Id] == nil {
			clients[*account.Id] = map[string]*cloudformation.Client{}
		}

		for _, region := range regions {
			clients[*account.Id][region] = cloudformation.NewFromConfig(cfg, func(o *cloudformation.Options) { o.Region = region })
		}
	}

	return clients, nil
}

// Get Accounts for specific Organizational Units. Taken from official AWS
// plugin verbatim, see original licence here:
// https://github.com/cloudquery/cloudquery/blob/7f0128f2ba1af1cb88cfa4de93cfee148959c488/LICENSE
// which also applies to this code.
func getOUAccounts(ctx context.Context, orgClient *organizations.Client, org *AwsOrg) ([]types.Account, error) {
	q := org.OrganizationUnits
	var ou string
	var rawAccounts []types.Account
	seenOUs := map[string]struct{}{}
	for len(q) > 0 {
		ou, q = q[0], q[1:]

		// Skip duplicates to avoid making duplicate API calls
		if _, found := seenOUs[ou]; found {
			continue
		}
		seenOUs[ou] = struct{}{}

		// get accounts directly under this OU
		accountsPaginator := organizations.NewListAccountsForParentPaginator(orgClient, &organizations.ListAccountsForParentInput{
			ParentId: aws.String(ou),
		})
		for accountsPaginator.HasMorePages() {
			output, err := accountsPaginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			rawAccounts = append(rawAccounts, output.Accounts...)
		}

		// get OUs directly under this OU, and add them to the queue
		ouPaginator := organizations.NewListChildrenPaginator(orgClient, &organizations.ListChildrenInput{
			ChildType: types.ChildTypeOrganizationalUnit,
			ParentId:  aws.String(ou),
		})
		for ouPaginator.HasMorePages() {
			output, err := ouPaginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			for _, child := range output.Children {
				q = append(q, *child.Id)
			}
		}
	}

	return rawAccounts, nil
}

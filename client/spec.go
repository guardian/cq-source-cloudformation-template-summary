package client

// These are subsets of the official AWS plugin.
type Account struct {
	ID           string `json:"id"`
	LocalProfile string `json:"local_profile,omitempty"`
}

type AwsOrg struct {
	OrganizationUnits    []string `json:"organization_units,omitempty"`
	ChildAccountRoleName string   `json:"member_role_name,omitempty"`
}
type Spec struct {
	Accounts     []Account `json:"accounts"`
	Organization *AwsOrg   `json:"org"`
	Regions      []string  `json:"regions,omitempty"`
}

package client

type AffiliationsResponse struct {
	Affiliations Affiliation `json:"affiliations"`
}

type Secret struct {
	Name          string `json:"name"`
	Base64Content string `json:"base64Content"`
}

type Vault struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	HasAccess   bool     `json:"hasAccess"`
	Secrets     []Secret `json:"secrets"`
}

type Node struct {
	Name   string  `json:"name"`
	Vaults []Vault `json:"vaults"`
}

type Edge struct {
	Node Node `json:"node"`
}

type Affiliation struct {
	Edges []Edge `json:"edges"`
}

// NewVault creates a new Vault (illegal, since it is missing both secrets and permissions)
func NewVault(name string) *Vault {
	return &Vault{
		Name:        name,
		Secrets:     []Secret{},
		Permissions: []string{},
	}
}

func (api *AffiliationsResponse) Vaults(affiliation string) []Vault {
	for _, edge := range api.Affiliations.Edges {
		if edge.Node.Name == affiliation {
			return edge.Node.Vaults
		}
	}
	return nil
}

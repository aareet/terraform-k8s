package workspace

import (
	"context"
	"fmt"
	"net/url"
	"os"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/command/cliconfig"
)

var (
	// AutoApply run to workspace
	AutoApply = true
)

// TerraformCloudClient has a TFC Client and organization
type TerraformCloudClient struct {
	Client           *tfc.Client
	Organization     string
	SecretsMountPath string
}

func createTerraformConfig(tfConfig *cliconfig.Config) (*tfc.Config, error) {
	address := os.Getenv("TF_URL")
	if len(address) == 0 {
		address = tfc.DefaultAddress
	}
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("Not a valid TF URL, %v", err)
	}
	host := u.Host

	if len(tfConfig.Credentials[host]) == 0 {
		fmt.Println(tfConfig.Credentials[host])
		return nil, fmt.Errorf("Define token for %s", host)
	}

	return &tfc.Config{
		Address: address,
		Token:   fmt.Sprintf("%v", tfConfig.Credentials[host]["token"]),
	}, nil
}

// GetClient creates the configuration for Terraform Cloud
func (t *TerraformCloudClient) GetClient() error {
	tfConfig, diag := cliconfig.LoadConfig()
	if diag.Err() != nil {
		return diag.Err()
	}

	config, err := createTerraformConfig(tfConfig)
	if err != nil {
		return err
	}

	client, err := tfc.NewClient(config)
	if err != nil {
		return diag.Err()
	}

	t.Client = client
	return nil
}

// CheckOrganization looks for an organization
func (t *TerraformCloudClient) CheckOrganization() error {
	_, err := t.Client.Organizations.Read(context.TODO(), t.Organization)
	return err
}

// CheckWorkspace looks for a workspace
func (t *TerraformCloudClient) CheckWorkspace(workspace string) (string, error) {
	ws, err := t.Client.Workspaces.Read(context.TODO(), t.Organization, workspace)
	if err != nil && err == tfc.ErrResourceNotFound {
		id, err := t.CreateWorkspace(workspace)
		if err != nil {
			return "", err
		}
		return id, nil
	} else if err != nil {
		return "", err
	}
	return ws.ID, err
}

// CreateWorkspace creates a Terraform Cloud Workspace that auto-applies
func (t *TerraformCloudClient) CreateWorkspace(workspace string) (string, error) {
	options := tfc.WorkspaceCreateOptions{
		AutoApply: &AutoApply,
		Name:      &workspace,
	}
	ws, err := t.Client.Workspaces.Create(context.TODO(), t.Organization, options)
	if err != nil {
		return "", err
	}
	return ws.ID, nil
}

// CheckWorkspacebyID checks a workspace by ID
func (t *TerraformCloudClient) CheckWorkspacebyID(workspaceID string) error {
	_, err := t.Client.Workspaces.ReadByID(context.TODO(), workspaceID)
	return err
}

// DeleteWorkspace removes the workspace from Terraform Cloud
func (t *TerraformCloudClient) DeleteWorkspace(workspaceID string) error {
	err := t.Client.Workspaces.DeleteByID(context.TODO(), workspaceID)
	if err != nil {
		return err
	}
	return nil
}

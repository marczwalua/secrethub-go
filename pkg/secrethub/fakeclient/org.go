// +build !production

package fakeclient

import (
	"github.com/secrethub/secrethub-go/internals/api"
	"github.com/secrethub/secrethub-go/pkg/secrethub"
)

// OrgService is a mock of the RepoService interface.
type OrgService struct {
	Creater       OrgCreater
	Deleter       OrgDeleter
	Getter        OrgGetter
	MemberService *OrgMemberService
	MineLister    OrgMineLister
}

// Create implements the RepoService interface Create function.
func (s *OrgService) Create(name string, description string) (*api.Org, error) {
	return s.Creater.Create(name, description)
}

// Delete implements the RepoService interface Delete function.
func (s *OrgService) Delete(name string) error {
	return s.Deleter.Delete(name)
}

// Get implements the RepoService interface Get function.
func (s *OrgService) Get(name string) (*api.Org, error) {
	return s.Getter.Get(name)
}

// Members returns a mock of the OrgMemberService interface.
func (s *OrgService) Members() secrethub.OrgMemberService {
	return s.MemberService
}

// ListMine implements the RepoService interface ListMine function.
func (s *OrgService) ListMine() ([]*api.Org, error) {
	return s.MineLister.ListMine()
}

// OrgCreater mocks the Create function.
type OrgCreater struct {
	ArgName        string
	ArgDescription string
	ReturnsOrg     *api.Org
	Err            error
}

// Create saves the arguments it was called with and returns the mocked response.
func (c *OrgCreater) Create(name string, description string) (*api.Org, error) {
	c.ArgName = name
	c.ArgDescription = description
	return c.ReturnsOrg, c.Err
}

// OrgDeleter mocks the Delete function.
type OrgDeleter struct {
	ArgName string
	Err     error
}

// Delete saves the arguments it was called with and returns the mocked response.
func (d *OrgDeleter) Delete(name string) error {
	d.ArgName = name
	return d.Err
}

// OrgGetter mocks the Get function.
type OrgGetter struct {
	ArgName    string
	ReturnsOrg *api.Org
	Err        error
}

// Get saves the arguments it was called with and returns the mocked response.
func (g *OrgGetter) Get(name string) (*api.Org, error) {
	g.ArgName = name
	return g.ReturnsOrg, g.Err
}

// OrgMineLister mocks the ListMine function.
type OrgMineLister struct {
	ReturnsOrgs []*api.Org
	Err         error
}

// ListMine returns the mocked response.
func (m *OrgMineLister) ListMine() ([]*api.Org, error) {
	return m.ReturnsOrgs, m.Err
}

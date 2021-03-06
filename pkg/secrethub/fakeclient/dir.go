// +build !production

package fakeclient

import (
	"github.com/secrethub/secrethub-go/internals/api"
)

// DirService is a mock of the DirService interface.
type DirService struct {
	Creater    DirCreater
	Deleter    DirDeleter
	TreeGetter TreeGetter
}

// Create implements the DirService interface Create function.
func (s *DirService) Create(path string) (*api.Dir, error) {
	return s.Creater.Create(path)
}

// Delete implements the DirService interface Delete function.
func (s *DirService) Delete(path string) error {
	return s.Deleter.Delete(path)
}

// GetTree implements the DirService interface GetTree function.
func (s *DirService) GetTree(path string, depth int, ancestors bool) (*api.Tree, error) {
	return s.TreeGetter.GetTree(path, depth)
}

// DirCreater mocks the Create function.
type DirCreater struct {
	ArgPath    string
	ReturnsDir *api.Dir
	Err        error
}

// Create saves the arguments it was called with and returns the mocked response.
func (dc *DirCreater) Create(path string) (*api.Dir, error) {
	dc.ArgPath = path
	return dc.ReturnsDir, dc.Err
}

// DirDeleter mocks the Delete function.
type DirDeleter struct {
	ArgPath string
	Err     error
}

// Delete saves the arguments it was called with and returns the mocked response.
func (d *DirDeleter) Delete(path string) error {
	d.ArgPath = path
	return d.Err
}

// TreeGetter mocks the Get function.
type TreeGetter struct {
	ArgPath     string
	ArgDepth    int
	ReturnsTree *api.Tree
	Err         error
}

// GetTree saves the arguments it was called with and returns the mocked response.
func (dg *TreeGetter) GetTree(path string, depth int) (*api.Tree, error) {
	dg.ArgPath = path
	dg.ArgDepth = depth
	return dg.ReturnsTree, dg.Err
}

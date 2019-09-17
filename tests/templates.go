package tests

import (
	"github.com/SebastienDorgan/talgo"
	"reflect"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

//TemplateManagerTestSuite test suite off api.Temp
type TemplateManagerTestSuite struct {
	suite.Suite
	Mgr api.ServerTemplateManager
}

func checkDouble(templates []api.ServerTemplate) bool {
	size := len(templates)
	n := talgo.FindFirst(size, func(i int) bool {
		tail := templates[i+1 : size]
		n := talgo.FindFirst(len(tail), func(j int) bool {
			return tail[j].ID == templates[i].ID
		})
		return n < len(tail) && n > 0
	})
	return n < size && n > 0
}

//TestServerTemplateManager Canonical test for ServerTemplateManager implementation
func (s *TemplateManagerTestSuite) TestServerTemplateManager() {
	mgr := s.Mgr
	templates, err := mgr.List()
	assert.NoError(s.T(), err)
	assert.False(s.T(), checkDouble(templates))
	assert.True(s.T(), len(templates) > 0)
	for _, tpl := range templates {
		tp, err := mgr.Get(tpl.ID)
		assert.NoError(s.T(), err)
		assert.True(s.T(), reflect.DeepEqual(tpl, *tp))
		assert.NotEmpty(s.T(), tp.Name)
		assert.NotEqualf(s.T(), 0, tp.NumberOfCPUCore, "No CPU core for %s", tp.Name)
		assert.NotEqualf(s.T(), 0, tp.RAMSize, "No Memory for %s", tp.Name)
	}
}

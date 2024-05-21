package gone_grpc

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServerPriest(t *testing.T) {
	cemetery := gone.NewBuryMockCemeteryForTest()
	err := ServerPriest(cemetery)
	assert.Nil(t, err)
	err = ClientRegisterPriest(cemetery)
	assert.Nil(t, err)
}

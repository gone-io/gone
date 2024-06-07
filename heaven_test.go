package gone

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_heaven_burial(t *testing.T) {
	executed := false
	func() {
		defer func() {
			err := recover()
			assert.Equal(t, errors.New("test"), err)
			executed = true
		}()

		New(func(cemetery Cemetery) error {
			return errors.New("test")
		}).(*heaven).burial()
	}()
	assert.True(t, executed)
}

func Test_heaven_install(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCemetery := NewMockCemetery(ctrl)

	mockCemetery.EXPECT().ReviveAllFromTombs().Return(errors.New("ReviveAllFromTombs failed"))

	h := heaven{
		cemetery: mockCemetery,
	}

	executed := false
	func() {
		defer func() {
			err := recover()
			assert.Equal(t, errors.New("ReviveAllFromTombs failed"), err)
			executed = true
		}()
		h.install()
	}()
	assert.True(t, executed)
}

func Test_heaven_installAngelHook(t *testing.T) {
	h := New(func(cemetery Cemetery) error {
		cemetery.Bury(&angel{})
		return nil
	}).(*heaven)
	h.burial()
	h.installAngelHook()
	assert.Equal(t, 1, len(h.beforeStartHandlers))
	assert.Equal(t, 1, len(h.beforeStopHandlers))
}

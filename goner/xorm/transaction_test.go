package xorm

import (
	"errors"
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"xorm.io/xorm"
)

func Test_session(t *testing.T) {
	gone.Prepare(func(cemetery gone.Cemetery) error {
		_ = config.Priest(cemetery)
		_ = logrus.Priest(cemetery)
		return nil
	}).AfterStart(func(in struct {
		logger   gone.Logger   `gone:"gone-logger"`
		cemetery gone.Cemetery `gone:"gone-cemetery"`
	}) {
		t.Run("suc", func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			engineInterface := NewMockEngineInterface(controller)

			e := wrappedEngine{
				log: in.logger,
				newFunc: func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
					return engineInterface, nil
				},
				newSession: func(engineInterface xorm.EngineInterface) XInterface {
					session := NewMockXInterface(controller)
					session.EXPECT().Begin().Return(nil)
					session.EXPECT().Close().Return(nil)
					session.EXPECT().Commit().Return(nil)

					return session
				},
			}

			err := e.Transaction(func(session1 Interface) error {
				return e.Transaction(func(session2 Interface) error {
					assert.Equal(t, session1, session2)
					return nil
				})
			})

			assert.NoError(t, err)
		})

		t.Run("suc but Close error", func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			engineInterface := NewMockEngineInterface(controller)

			e := wrappedEngine{
				log: in.logger,
				newFunc: func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
					return engineInterface, nil
				},
				newSession: func(engineInterface xorm.EngineInterface) XInterface {
					session := NewMockXInterface(controller)
					session.EXPECT().Begin().Return(nil)
					session.EXPECT().Close().Return(errors.New("error"))
					session.EXPECT().Commit().Return(nil)

					return session
				},
			}

			err := e.Transaction(func(session1 Interface) error {
				return e.Transaction(func(session2 Interface) error {
					assert.Equal(t, session1, session2)
					return nil
				})
			})

			assert.NoError(t, err)
		})

		t.Run("rollback", func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			engineInterface := NewMockEngineInterface(controller)

			e := wrappedEngine{
				log: in.logger,
				newFunc: func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
					return engineInterface, nil
				},
				newSession: func(engineInterface xorm.EngineInterface) XInterface {
					session := NewMockXInterface(controller)
					session.EXPECT().Begin().Return(nil)
					session.EXPECT().Close().Return(nil)
					session.EXPECT().Rollback().Return(nil)

					return session
				},
			}

			err := e.Transaction(func(session1 Interface) error {
				return e.Transaction(func(session2 Interface) error {
					assert.Equal(t, session1, session2)
					return errors.New("error")
				})
			})

			assert.Error(t, err)
		})

		t.Run("begin error", func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			engineInterface := NewMockEngineInterface(controller)

			e := wrappedEngine{
				log: in.logger,
				newFunc: func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
					return engineInterface, nil
				},
				newSession: func(engineInterface xorm.EngineInterface) XInterface {
					session := NewMockXInterface(controller)
					session.EXPECT().Begin().Return(errors.New("error"))
					session.EXPECT().Close().Return(nil)

					return session
				},
			}

			err := e.Transaction(func(session1 Interface) error {
				return e.Transaction(func(session2 Interface) error {
					assert.Equal(t, session1, session2)
					return errors.New("error")
				})
			})

			assert.Error(t, err)
		})

		t.Run("roll panic", func(t *testing.T) {
			executed := false
			func() {
				controller := gomock.NewController(t)
				defer controller.Finish()

				defer func() {
					err := recover()
					if err != nil {
						executed = true
					}
				}()

				engineInterface := NewMockEngineInterface(controller)

				e := wrappedEngine{
					log: in.logger,
					newFunc: func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
						return engineInterface, nil
					},
					newSession: func(engineInterface xorm.EngineInterface) XInterface {
						session := NewMockXInterface(controller)
						session.EXPECT().Begin().Return(nil)
						session.EXPECT().Close().Return(nil)
						session.EXPECT().Rollback().Do(func() {
							panic("error")
						}).Return(nil)

						return session
					},
				}

				_ = e.Transaction(func(session1 Interface) error {
					return e.Transaction(func(session2 Interface) error {
						assert.Equal(t, session1, session2)
						return errors.New("error")
					})
				})
			}()

			assert.True(t, executed)
		})

		t.Run("roll error", func(t *testing.T) {
			func() {
				controller := gomock.NewController(t)
				defer controller.Finish()

				engineInterface := NewMockEngineInterface(controller)

				e := wrappedEngine{
					log: in.logger,
					newFunc: func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
						return engineInterface, nil
					},
					newSession: func(engineInterface xorm.EngineInterface) XInterface {
						session := NewMockXInterface(controller)
						session.EXPECT().Begin().Return(nil)
						session.EXPECT().Close().Return(nil)
						session.EXPECT().Rollback().Return(errors.New("error"))

						return session
					},
				}

				err := e.Transaction(func(session1 Interface) error {
					return e.Transaction(func(session2 Interface) error {
						assert.Equal(t, session1, session2)
						return errors.New("error")
					})
				})
				assert.Error(t, err)
			}()
		})

		t.Run("panic in transaction func", func(t *testing.T) {
			func() {
				controller := gomock.NewController(t)
				defer controller.Finish()

				engineInterface := NewMockEngineInterface(controller)

				e := wrappedEngine{
					log: in.logger,
					newFunc: func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
						return engineInterface, nil
					},
					newSession: func(engineInterface xorm.EngineInterface) XInterface {
						session := NewMockXInterface(controller)
						session.EXPECT().Begin().Return(nil)
						session.EXPECT().Close().Return(nil)
						session.EXPECT().Rollback().Return(nil)

						return session
					},
				}

				err := e.Transaction(func(session1 Interface) error {
					return e.Transaction(func(session2 Interface) error {
						panic("error")
					})
				})
				assert.Error(t, err)
			}()

		})
	}).Run()
}

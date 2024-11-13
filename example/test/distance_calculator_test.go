package test

import (
	"example/test/mock"
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_distanceCalculator_CalculateDistanceFromOrigin(t *testing.T) {

	//创建mock控制器
	controller := gomock.NewController(t)
	defer controller.Finish()

	gone.Test(func(d *distanceCalculator) {
		distance := d.CalculateDistanceFromOrigin(3, 4)

		assert.Equal(t, float64(5), distance)

	}, func(cemetery gone.Cemetery) error {

		//创建mock对象
		point := mock.NewMockIPoint(controller)
		point.EXPECT().GetX().Return(0)
		point.EXPECT().GetY().Return(0)

		//将mock对象埋葬到Cemetery
		cemetery.Bury(point)

		//被测试的对象也需要埋葬到Cemetery
		cemetery.Bury(NewDistanceCalculator())
		return nil
	})
}

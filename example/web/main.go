package main

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
	"github.com/gone-io/gone/goner/gin"
)

type controller struct {
	gone.Flag
	router gin.IRouter `gone:"gone-gin-router"` //inject gin router Goner, which is wrapped of `gin.Engine`
}

// Mount use for  mounting the router of gin framework
func (ctr *controller) Mount() gin.MountError {
	ctr.router.GET("/ping", func(c *gin.Context) (any, error) {
		return "hello", nil
	})
	return nil
}

func NewController() gone.Goner {
	return &controller{}
}

func Priest(cemetery gone.Cemetery) error {
	//Load the Goner of the gin web framework into the system
	_ = goner.GinPriest(cemetery)

	//Load the business Goner
	cemetery.Bury(NewController())
	return nil
}

func main() {

	//Gone.Server is used to start a service, and the program will block until the service ends.
	gone.Serve(Priest)
}

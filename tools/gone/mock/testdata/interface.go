package testdata

//go:generate sh -c "mockgen -source=interface.go -package=testdata  -destination=testInterface.go"
type TestInterface interface {
	GetSomeThing() error
	DoSomeThing(bool bool) (string, error)
}

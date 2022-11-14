package testdata

//go:generate sh -c "mockgen -package=mock -source=$GOFILE | ../../bin/gone mock -o mock/$GOFILE"
type TestInterface interface {
	GetSomeThing() error
	DoSomeThing(bool bool) (string, error)
}

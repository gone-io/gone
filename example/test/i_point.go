package test

//go:generate sh -c "mockgen -package=mock -source=$GOFILE|gone mock -o mock/$GOFILE"
type IPoint interface {
	GetX() int
	GetY() int
}

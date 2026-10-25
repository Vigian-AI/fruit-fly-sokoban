package game

type Position struct{ Row, Col int }

type State struct {
	Player Position
	Boxes  []Position
}

package example

type foo struct {
	N int
}

// Next returns the next value in sequence.
func (f *foo) Next() int {
	x := f.Peek()
	f.N++
	return x
}

// Peek returns the next value without updating the seqence.
func (f foo) Peek() int {
	return f.N
}

// Reset sets the sequence to start at the given value
func (f *foo) Reset(n int) {
	f.N = n
}

//go:generate go run github.com/cookieo9/packager -local defaultFoo -block Init
//go:generate go run github.com/cookieo9/packager -local defaultFoo -output -
var defaultFoo foo

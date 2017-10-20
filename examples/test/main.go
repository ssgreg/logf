package main

type Greg interface {
	Foo(f string, v interface{}) Greg
}

type GregImpl struct {
}

func (g *GregImpl) Foo(f string, v interface{}) Greg {
	println(v)
	return g
}

func main() {
	l := "test"//make([]int, 10, 10)
	g := GregImpl{}
	g.Foo("footest", l).Foo("bartest", l)
}

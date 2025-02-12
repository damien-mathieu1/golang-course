package main

import "fmt"

type vec2i struct {
	x, y int
}

func (v vec2i) initVec2i(x, y int) {
	v.x = x
	v.y = y
}

func (v vec2i) add(v2 vec2i) vec2i {
	return vec2i{v.x + v2.x, v.y + v2.y}
}

func (v vec2i) sub(v2 vec2i) vec2i {
	return vec2i{v.x - v2.x, v.y - v2.y}
}

func main() {

	fmt.Println("Hello, World!")
}

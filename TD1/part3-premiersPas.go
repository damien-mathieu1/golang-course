package main

import "fmt"

type vec2i struct {
	x, y int
}

func (v *vec2i) initVec2i(x, y int) {
	v.x = x
	v.y = y
}

func (v vec2i) add(v2 vec2i) vec2i {
	return vec2i{v.x + v2.x, v.y + v2.y}
}

func (v vec2i) sub(v2 vec2i) vec2i {
	return vec2i{v.x - v2.x, v.y - v2.y}
}

func (v vec2i) multiply(v2 vec2i) vec2i {
	return vec2i{v.x * v2.x, v.y * v2.y}
}

func (v vec2i) norme() int {
	return v.x*v.x + v.y*v.y
}

func (v vec2i) normalize() vec2i {
	return vec2i{v.x / v.norme(), v.y / v.norme()}
}

func (v vec2i) scalar(v2 vec2i) int {
	return v.x*v2.x + v.y*v2.y
}

func (v vec2i) produitVectoriel(v2 vec2i) int {
	return v.x*v2.y - v.y*v2.x
}
func main() {
	var v1, v2 vec2i
	v1.initVec2i(13, 245)
	v2.initVec2i(321, 42)

	fmt.Println(v1.add(v2))
	fmt.Println(v1.sub(v2))
	fmt.Println(v1.multiply(v2))
	fmt.Println(v1.norme())
	fmt.Println(v1.normalize())
	fmt.Println(v1.scalar(v2))
	fmt.Println(v1.produitVectoriel(v2))
}

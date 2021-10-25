package goOctree

import "math"

type Vector3 struct {
	X float32
	Y float32
	Z float32
}

func (v Vector3) Add(v2 Vector3) Vector3 {
	v.X += v2.X
	v.Y += v2.Y
	v.Z += v2.Z

	return v
}

func (v Vector3) Multiply(num float32) Vector3 {
	v.X *= num
	v.Y *= num
	v.Z *= num

	return v
}

func Distance(a, b Vector3) float32 {
	dist := math.Sqrt(math.Pow(float64(a.X-b.X), 2) + math.Pow(float64(a.Y-b.Y), 2) + math.Pow(float64(a.Z-b.Z), 2))
	return float32(dist)
}

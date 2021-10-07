package goOctree

type vector3 struct {
	x float32
	y float32
	z float32
}

func (v vector3) Add(v2 vector3) vector3 {
	v.x += v2.x
	v.y += v2.y
	v.z += v2.z

	return v
}

func (v vector3) Multiply(num float32) vector3 {
	v.x *= num
	v.y *= num
	v.z *= num

	return v
}

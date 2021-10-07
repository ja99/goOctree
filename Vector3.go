package goOctree

type Vector3 struct {
	x float32
	y float32
	z float32
}

func (v Vector3) Add(v2 Vector3) Vector3 {
	v.x += v2.x
	v.y += v2.y
	v.z += v2.z

	return v
}

func (v Vector3) Multiply(num float32) Vector3 {
	v.x *= num
	v.y *= num
	v.z *= num

	return v
}

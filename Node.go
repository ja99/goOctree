package goOctree

import (
	"strconv"
)

type Node struct {
	Uid    string
	Center Vector3
	Size   float32
	// [0] = -X -Y -Z //left low back
	// [1] = -X -Y +Z //left low front
	// [2] = -X +Y -Z //left high back
	// [3] = -X +Y +Z //left high front
	// [4] = +X -Y -Z //right low back
	// [5] = +X -Y +Z //right low front
	// [6] = +X +Y -Z //right high back
	// [7] = +X +Y +Z //right high front
	Children [8]*Node
	Point    *Vector3
	MaxDepth uint8
	Parent   *Node
}

func (n *Node) MakeChildren() {
	counter := 0
	for x := float32(-1); x <= 1; x += 2 {
		for y := float32(-1); y <= 1; y += 2 {
			for z := float32(-1); z <= 1; z += 2 {
				nudge := Vector3{
					X: n.Size * 0.25 * x,
					Y: n.Size * 0.25 * y,
					Z: n.Size * 0.25 * z,
				}
				newCenter := n.Center.Add(nudge)

				newNode := Node{
					Uid:      n.Uid + strconv.Itoa(counter),
					Center:   newCenter,
					Size:     n.Size * 0.5,
					Children: [8]*Node{},
					Point:    nil,
					MaxDepth: 0,
					Parent:   n,
				}
				n.Children[counter] = &newNode

				counter++
			}
		}
	}
	n.RaiseMaxDepth(n.MaxDepth + 1)
}

func (n *Node) RaiseMaxDepth(childDepth uint8) {

	if childDepth+1 > n.MaxDepth {
		n.MaxDepth = childDepth + 1

		if n.Parent != nil {
			n.Parent.RaiseMaxDepth(n.MaxDepth)
		}
	}
}

func inside(low float32, high float32, val float32) bool {
	return (low <= val) && (val <= high)
}

func (n *Node) PointFits(point *Vector3) bool {
	if !inside(n.Center.X-0.5*n.Size, n.Center.X+0.5*n.Size, point.X) {
		return false
	}
	if !inside(n.Center.Y-0.5*n.Size, n.Center.Y+0.5*n.Size, point.Y) {
		return false
	}
	if !inside(n.Center.Z-0.5*n.Size, n.Center.Z+0.5*n.Size, point.Z) {
		return false
	}
	return true
}

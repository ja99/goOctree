package goOctree

import (
	"strconv"
	"sync"
)

type Node struct {
	uid    string
	center Vector3
	size   float32
	// [0] = -x -y -z //left low back
	// [1] = -x -y +z //left low front
	// [2] = -x +y -z //left high back
	// [3] = -x +y +z //left high front
	// [4] = +x -y -z //right low back
	// [5] = +x -y +z //right low front
	// [6] = +x +y -z //right high back
	// [7] = +x +y +z //right high front
	children [8]*Node
	point    *Vector3
	maxDepth uint8
	parent   *Node
	lock     sync.Mutex
}

func (n *Node) MakeChildren() {
	counter := 0
	for x := float32(-1); x <= 1; x += 2 {
		for y := float32(-1); y <= 1; y += 2 {
			for z := float32(-1); z <= 1; z += 2 {
				nudge := Vector3{
					x: n.size * 0.25 * x,
					y: n.size * 0.25 * y,
					z: n.size * 0.25 * z,
				}
				newCenter := n.center.Add(nudge)

				newNode := Node{
					uid:      n.uid + strconv.Itoa(counter),
					center:   newCenter,
					size:     n.size * 0.5,
					children: [8]*Node{},
					point:    nil,
					maxDepth: 0,
					parent:   n,
					lock:     sync.Mutex{},
				}
				n.children[counter] = &newNode

				counter++
			}
		}
	}
	n.RaiseMaxDepth(n.maxDepth + 1)
}

func (n *Node) RaiseMaxDepth(childDepth uint8) {

	if childDepth+1 > n.maxDepth {
		n.maxDepth = childDepth + 1

		if n.parent != nil {
			n.parent.RaiseMaxDepth(n.maxDepth)
		}
	}
}

func Inside(low float32, high float32, val float32) bool {
	return (low <= val) && (val <= high)
}

func (n *Node) PointFits(point *Vector3) bool {
	if !Inside(n.center.x-0.5*n.size, n.center.x+0.5*n.size, point.x) {
		return false
	}
	if !Inside(n.center.y-0.5*n.size, n.center.y+0.5*n.size, point.y) {
		return false
	}
	if !Inside(n.center.z-0.5*n.size, n.center.z+0.5*n.size, point.z) {
		return false
	}
	return true
}

package goOctree

import (
	"errors"
	"fmt"
	"math"
	"sync"
)

type Octree struct {
	Root *Node
}

var (
	aitErr = errors.New("Point was already in tree")
	stmErr = errors.New("Node was smaller than minSize")
	nfsErr = errors.New("No free Space found")
	nhnErr = errors.New("NEW HOME WAS STILL NIL")
	nffErr = errors.New("DID NOT FIND A FITTING CHILD")
	nfnErr = errors.New("DID NOT FIND A FITTING NODE")
)

// Insertion
func (tree *Octree) Insert(point *Vector3, minSize float32, verbose bool) []*Node {
	var createdNodes []*Node

	newHome, err := FindFreeSpace(tree.Root, point, minSize, &createdNodes)
	if err != nil {
		if verbose {
			fmt.Println(err)
		}
		return createdNodes
	} else {
		newHome.Point = point
	}
	return createdNodes
}

func FindFreeSpace(currentNode *Node, point *Vector3, minSize float32, createdNodes *[]*Node) (*Node, error) {
	//If this is true, we have found the fitting node
	if currentNode.IsFree() {
		//if currentNode.Point == nil && currentNode.MaxDepth == 0 {
		return currentNode, nil
	}

	//Test If Point is already inside the Octree
	if currentNode.Point != nil {
		if *currentNode.Point == *point {
			return nil, aitErr
		}
	}

	//Test if node size is smaller than minSize
	if currentNode.Size < minSize {
		return nil, stmErr
	}

	//if you got til here, then we need to go further down
	//But if we don't have Children then we need to make them first
	if currentNode.Children[0] == nil {
		//Don`t make children if they would be too small
		if currentNode.Size*0.5 < minSize {
			return nil, stmErr
		}
		currentNode.MakeChildren()
		//Add Children to list of newly crated Nodes
		for _, child := range currentNode.Children {
			*createdNodes = append(*createdNodes, child)
		}
		// Since points are only allowed in child nodes, we will have to trickle down the Point
		v := currentNode.Point
		currentNode.Point = nil
		newHome, err := FindFreeSpace(currentNode, v, minSize, createdNodes)
		if err != nil {
			fmt.Println(err)
		} else if newHome == nil {
			return nil, nhnErr
		} else {
			newHome.Point = v
		}
	}

	//Look for candidate in Children
	for i := 0; i < 8; i++ {
		if currentNode.Children[i].PointFits(point) {
			return FindFreeSpace(currentNode.Children[i], point, minSize, createdNodes)
		}
	}

	return nil, nfsErr
}

//Enum for CheckPoint
type Direction int

const (
	xMinus Direction = iota
	xPlus            //1
	yMinus           //2
	yPlus            //3
	zMinus           //4
	zPlus            //5
	all              //6
)

// Points which should identify neighbours
type CheckPoint struct {
	Point *Vector3
	Dir   Direction
}

// Neighbor Query
func GetNeighbors(currentNode *Node, rootNode *Node, hasToBeFree bool, onlyDirectNeighbours bool) []*Node {
	// Make Points in each direction to find the neighbours
	checkPoints := []*CheckPoint{}
	for axis := 0; axis < 3; axis++ {
		for plusOrMinus := float32(-1); plusOrMinus < 2; plusOrMinus += 2 {
			p := &Vector3{}
			var dir Direction
			if axis == 0 {
				p = currentNode.Center.Add(&Vector3{
					X: plusOrMinus*currentNode.Size*0.5 + plusOrMinus*0.0001,
					Y: 0,
					Z: 0,
				})
			}
			if axis == 1 {
				p = currentNode.Center.Add(&Vector3{
					X: 0,
					Y: plusOrMinus*currentNode.Size*0.5 + plusOrMinus*0.0001,
					Z: 0,
				})
			}
			if axis == 2 {
				p = currentNode.Center.Add(&Vector3{
					X: 0,
					Y: 0,
					Z: plusOrMinus*currentNode.Size*0.5 + plusOrMinus*0.0001,
				})
			}
			if onlyDirectNeighbours {
				dir = Direction(axis*2 + int(math.Max(0, float64(-plusOrMinus))))
			} else {
				dir = all
			}

			checkPoint := CheckPoint{
				Point: p,
				Dir:   dir,
			}
			checkPoints = append(checkPoints, &checkPoint)
		}
	}

	var returnSlice []*Node
	for _, checkPoint := range checkPoints {
		if rootNode.PointFits(checkPoint.Point) {
			n, err := FindFittingChild(rootNode, checkPoint.Point, len(currentNode.Uid))
			if err != nil {
				fmt.Println(err)
				continue
			}

			var addition []*Node

			if n.Children[0] == nil {
				addition = append(addition, n)
			} else {
				addition = GetChildrenRecursively(n, hasToBeFree, checkPoint.Dir)
			}
			returnSlice = append(returnSlice, addition...)
		}
	}
	return returnSlice

}

func FindFittingChild(currentNode *Node, point *Vector3, depth int) (*Node, error) {
	if len(currentNode.Uid) == depth {
		return currentNode, nil

	}

	if currentNode.Children[0] == nil {
		return currentNode, nil
	}

	for _, child := range currentNode.Children {
		if child.PointFits(point) {
			return FindFittingChild(child, point, depth)
		}
	}

	return nil, nffErr
}

func GetChildrenRecursively(currentNode *Node, hasToBeFree bool, side Direction) []*Node {
	var returnSlice []*Node
	GetChildrenRecursivelyTask(currentNode, &returnSlice, hasToBeFree, side)
	return returnSlice
}

func GetChildrenRecursivelyTask(currentNode *Node, returnSlice *[]*Node, hasToBeFree bool, side Direction) {
	if currentNode.HasChildren() {
		if side == xMinus {
			for i := 0; i < 5; i++ {
				GetChildrenRecursivelyTask(currentNode.Children[i], returnSlice, hasToBeFree, side)
			}
		} else if side == xPlus {
			for i := 4; i < 8; i++ {
				GetChildrenRecursivelyTask(currentNode.Children[i], returnSlice, hasToBeFree, side)
			}
		} else if side == yMinus {
			for i := 0; i < 2; i++ {
				GetChildrenRecursivelyTask(currentNode.Children[i], returnSlice, hasToBeFree, side)
			}
			for i := 4; i < 6; i++ {
				GetChildrenRecursivelyTask(currentNode.Children[i], returnSlice, hasToBeFree, side)
			}
		} else if side == yPlus {
			for i := 2; i < 4; i++ {
				GetChildrenRecursivelyTask(currentNode.Children[i], returnSlice, hasToBeFree, side)
			}
			for i := 6; i < 8; i++ {
				GetChildrenRecursivelyTask(currentNode.Children[i], returnSlice, hasToBeFree, side)
			}
		} else if side == zMinus {
			for i := 0; i < 8; i += 2 {
				GetChildrenRecursivelyTask(currentNode.Children[i], returnSlice, hasToBeFree, side)
			}
		} else if side == zPlus {
			for i := 1; i < 8; i += 2 {
				GetChildrenRecursivelyTask(currentNode.Children[i], returnSlice, hasToBeFree, side)
			}
		} else if side == all {
			for i := 0; i < 8; i++ {
				GetChildrenRecursivelyTask(currentNode.Children[i], returnSlice, hasToBeFree, side)
			}
		} else {
			fmt.Println("ENUM DIRECTION OUTSIDE RANGE")
		}

	} else {
		if hasToBeFree {
			if currentNode.Point != nil {
				return
			}
		}

		*returnSlice = append(*returnSlice, currentNode)
	}
}

// ---------------------------------------------Unused Queries--------------------------------------------------------
func GetNodeWithUid(root *Node, uid string) (*Node, error) {
	currentNode := root
	for true {
		if currentNode.Uid == uid {
			return currentNode, nil
		}
		if len(currentNode.Uid) < len(uid) {
			for _, child := range currentNode.Children {
				if child.Uid == uid {
					currentNode = child
					break
				}
			}
		} else {
			return nil, nfnErr
		}
	}
	return nil, nfnErr
}

// Point already in tree query
func PointAlreadyInTree(currentNode *Node, point *Vector3) bool {
	//Test If Point is already inside the Octree
	if currentNode.Point != nil {
		//... Gentleman, we got him!
		if *currentNode.Point == *point {
			return true
		}
	}

	// Not in goal, but no more Children?
	// Then your princess is in another castle
	if currentNode.Children[0] == nil {
		return false
	}

	//Look for candidate in Children
	for i := 0; i < 8; i++ {
		if currentNode.Children[i].PointFits(point) {
			return PointAlreadyInTree(currentNode.Children[i], point)
		}
	}

	return false
}

// Spaces Query
type returnObjPoints struct {
	resultSlice []Vector3
	lock        sync.Mutex
}

func GetPoints(currentNode *Node) []Vector3 {
	//var resultSlice []Vector3
	wg := sync.WaitGroup{}

	ownChan := returnObjPoints{
		resultSlice: []Vector3{},
		lock:        sync.Mutex{},
	}

	wg.Add(1)
	go GetPointsTask(currentNode, &ownChan, &wg)
	wg.Wait()

	return ownChan.resultSlice
}

func GetPointsTask(currentNode *Node, ownchan *returnObjPoints, parentWg *sync.WaitGroup) {
	//var returnSlice []string
	defer parentWg.Done()
	//defer println(currentNode.Uid, "done")
	//fmt.Println("Uid", currentNode.Uid, "wg: ", parentWg)
	if currentNode.HasChildren() {
		parentWg.Add(8)
		for _, child := range currentNode.Children {
			go GetPointsTask(child, ownchan, parentWg)
		}

	} else if currentNode.Point != nil {
		ownchan.lock.Lock()
		ownchan.resultSlice = append(ownchan.resultSlice, *currentNode.Point)
		ownchan.lock.Unlock()
	}

}

// Spaces Query
type returnObjFreeSpaces struct {
	resultSlice []string
	lock        sync.Mutex
}

func GetFreeSpaces(currentNode *Node) []string {
	//var resultSlice []string
	wg := sync.WaitGroup{}

	ownChan := returnObjFreeSpaces{
		resultSlice: []string{},
		lock:        sync.Mutex{},
	}

	wg.Add(1)
	go GetFreeSpacesTask(currentNode, &ownChan, &wg)
	wg.Wait()

	return ownChan.resultSlice
}

func GetFreeSpacesTask(currentNode *Node, ownChan *returnObjFreeSpaces, parentWg *sync.WaitGroup) {
	//var returnSlice []string
	defer parentWg.Done()
	//defer println(currentNode.Uid, "done")
	//fmt.Println("Uid", currentNode.Uid, "wg: ", parentWg)
	if currentNode.HasChildren() {
		parentWg.Add(8)
		for _, child := range currentNode.Children {
			go GetFreeSpacesTask(child, ownChan, parentWg)
		}

	} else if currentNode.Point == nil {
		ownChan.lock.Lock()
		ownChan.resultSlice = append(ownChan.resultSlice, currentNode.Uid)
		ownChan.lock.Unlock()
	}

}

package goOctree

import (
	"errors"
	"fmt"
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

	//fmt.Println("Point:" , Point, " went to ", newHome.Uid)
}

func FindFreeSpace(currentNode *Node, point *Vector3, minSize float32, createdNodes *[]*Node) (*Node, error) {
	//If this is true...
	//... Gentleman, we got him!
	if currentNode.Point == nil && currentNode.Children[0] == nil {
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
		//fmt.Println("Point:" , v, " got moved from ", currentNode.Uid, " to ", newHome.Uid)
	}

	//Look for candidate in Children
	for i := 0; i < 8; i++ {
		if currentNode.Children[i].PointFits(point) {
			return FindFreeSpace(currentNode.Children[i], point, minSize, createdNodes)
		}
	}

	return nil, nfsErr
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
	if currentNode.Children[0] != nil {
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
	if currentNode.Children[0] != nil {
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

// Neighbor Query (All, not just directly facing)
//ToDo: MakeOnlyForDirectlayFacing option
// ToDo: Find a better solution for the "Check points"
// ToDo: Return NodePointers instead of Ids
func GetNeighbors(currentNode *Node, rootNode *Node, hasToBeFree bool) []string {
	checkPoints := []Vector3{}

	left := currentNode.Center.Add(Vector3{
		X: -currentNode.Size*0.5 - 0.0001,
		Y: 0,
		Z: 0,
	})
	checkPoints = append(checkPoints, left)

	right := currentNode.Center.Add(Vector3{
		X: currentNode.Size*0.5 + 0.0001,
		Y: 0,
		Z: 0,
	})

	checkPoints = append(checkPoints, right)

	bottom := currentNode.Center.Add(Vector3{
		X: 0,
		Y: -currentNode.Size*0.5 - 0.0001,
		Z: 0,
	})

	checkPoints = append(checkPoints, bottom)

	top := currentNode.Center.Add(Vector3{
		X: 0,
		Y: currentNode.Size*0.5 + 0.0001,
		Z: 0,
	})

	checkPoints = append(checkPoints, top)

	back := currentNode.Center.Add(Vector3{
		X: 0,
		Y: 0,
		Z: -currentNode.Size*0.5 - 0.0001,
	})

	checkPoints = append(checkPoints, back)

	front := currentNode.Center.Add(Vector3{
		X: 0,
		Y: 0,
		Z: currentNode.Size*0.5 + 0.0001,
	})
	checkPoints = append(checkPoints, front)

	var returnSlice []string

	for _, point := range checkPoints {
		if rootNode.PointFits(&point) {
			n, err := FindFittingChild(rootNode, &point, len(currentNode.Uid))
			if err != nil {
				fmt.Println(err)
				continue
			}

			addition := GetChildrenRecursively(n, hasToBeFree)
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

func GetChildrenRecursively(currentNode *Node, hasToBeFree bool) []string {
	var returnSlice []string
	GetChildrenRecursivelyTask(currentNode, &returnSlice, hasToBeFree)
	return returnSlice
}

func GetChildrenRecursivelyTask(currentNode *Node, returnSlice *[]string, hasToBeFree bool) {
	if currentNode.Children[0] != nil {
		for i := 0; i < 8; i++ {
			GetChildrenRecursivelyTask(currentNode.Children[i], returnSlice, hasToBeFree)
		}
	} else {
		if hasToBeFree {
			if currentNode.Point != nil {
				return
			}
		}

		*returnSlice = append(*returnSlice, currentNode.Uid)
	}
}

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

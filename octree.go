package goOctree

import (
	"fmt"
	"sync"
)

type octree struct {
	root *node
}

// Insertion
func (tree *octree) Insert(point *vector3) {
	newHome := FindFreeSpace(tree.root, point)
	if newHome == nil {
		fmt.Println("point", point, "was already in the tree")
		return
	}

	newHome.point = point

	//fmt.Println("point:" , point, " went to ", newHome.uid)
}

func FindFreeSpace(currentNode *node, point *vector3) *node {
	//If this is true...
	//... Gentleman, we got him!
	if currentNode.point == nil && currentNode.children[0] == nil {
		//if currentNode.point == nil && currentNode.maxDepth == 0 {
		return currentNode
	}

	//Test If point is already inside the octree
	if currentNode.point != nil {
		if *currentNode.point == *point {
			return nil
		}
	}

	//if you got til here, then we need to go further down
	//But if we don't have children then we need to make em first
	if currentNode.children[0] == nil {
		// Since points are only allowed in child nodes, we will have to trickle down the point
		currentNode.MakeChildren()
		v := currentNode.point
		currentNode.point = nil
		newHome := FindFreeSpace(currentNode, v)
		newHome.point = v
		//fmt.Println("point:" , v, " got moved from ", currentNode.uid, " to ", newHome.uid)
	}

	//Look for candidate in children
	for i := 0; i < 8; i++ {
		if currentNode.children[i].PointFits(point) {
			return FindFreeSpace(currentNode.children[i], point)
		}
	}

	fmt.Println("----------------------\n", "didnt find a free space!!!! \n----------------------")
	return nil
}

// Point already in tree query
func PointAlreadyInTree(currentNode *node, point *vector3) bool {
	//Test If point is already inside the octree
	if currentNode.point != nil {
		//... Gentleman, we got him!
		if *currentNode.point == *point {
			return true
		}
	}

	// Not in goal, but no more children?
	// Then your princess is in another castle
	if currentNode.children[0] == nil {
		return false
	}

	//Look for candidate in children
	for i := 0; i < 8; i++ {
		if currentNode.children[i].PointFits(point) {
			return PointAlreadyInTree(currentNode.children[i], point)
		}
	}

	return false
}

// Spaces Query
type returnObjPoints struct {
	resultSlice []vector3
	lock        sync.Mutex
}

func GetPoints(currentNode *node) []vector3 {
	//var resultSlice []vector3
	wg := sync.WaitGroup{}

	ownChan := returnObjPoints{
		resultSlice: []vector3{},
		lock:        sync.Mutex{},
	}

	wg.Add(1)
	go GetPointsTask(currentNode, &ownChan, &wg)
	wg.Wait()

	return ownChan.resultSlice
}

func GetPointsTask(currentNode *node, ownchan *returnObjPoints, parentWg *sync.WaitGroup) {
	//var returnSlice []string
	defer parentWg.Done()
	//defer println(currentNode.uid, "done")
	//fmt.Println("uid", currentNode.uid, "wg: ", parentWg)
	if currentNode.children[0] != nil {
		parentWg.Add(8)
		for _, child := range currentNode.children {
			go GetPointsTask(child, ownchan, parentWg)
		}

	} else if currentNode.point != nil {
		ownchan.lock.Lock()
		ownchan.resultSlice = append(ownchan.resultSlice, *currentNode.point)
		ownchan.lock.Unlock()
	}

}

// Spaces Query
type returnObjFreeSpaces struct {
	resultSlice []string
	lock        sync.Mutex
}

func GetFreeSpaces(currentNode *node) []string {
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

func GetFreeSpacesTask(currentNode *node, ownChan *returnObjFreeSpaces, parentWg *sync.WaitGroup) {
	//var returnSlice []string
	defer parentWg.Done()
	//defer println(currentNode.uid, "done")
	//fmt.Println("uid", currentNode.uid, "wg: ", parentWg)
	if currentNode.children[0] != nil {
		parentWg.Add(8)
		for _, child := range currentNode.children {
			go GetFreeSpacesTask(child, ownChan, parentWg)
		}

	} else if currentNode.point == nil {
		ownChan.lock.Lock()
		ownChan.resultSlice = append(ownChan.resultSlice, currentNode.uid)
		ownChan.lock.Unlock()
	}

}

// Neighbor Query (All, not just directly facing) ToDo: MakeOneOnlyForDirectlay Facing
func GetNeighbors(currentNode *node, rootNode *node, hasToBeFree bool) []string {
	checkPoints := []vector3{}

	left := currentNode.center.Add(vector3{
		x: -currentNode.size*0.5 - 0.0001,
		y: 0,
		z: 0,
	})
	checkPoints = append(checkPoints, left)

	right := currentNode.center.Add(vector3{
		x: currentNode.size*0.5 + 0.0001,
		y: 0,
		z: 0,
	})

	checkPoints = append(checkPoints, right)

	bottom := currentNode.center.Add(vector3{
		x: 0,
		y: -currentNode.size*0.5 - 0.0001,
		z: 0,
	})

	checkPoints = append(checkPoints, bottom)

	top := currentNode.center.Add(vector3{
		x: 0,
		y: currentNode.size*0.5 + 0.0001,
		z: 0,
	})

	checkPoints = append(checkPoints, top)

	back := currentNode.center.Add(vector3{
		x: 0,
		y: 0,
		z: -currentNode.size*0.5 - 0.0001,
	})

	checkPoints = append(checkPoints, back)

	front := currentNode.center.Add(vector3{
		x: 0,
		y: 0,
		z: currentNode.size*0.5 + 0.0001,
	})

	checkPoints = append(checkPoints, front)

	var returnSlice []string

	for _, point := range checkPoints {
		if rootNode.PointFits(&point) {
			n := FindFittingChild(rootNode, &point, len(currentNode.uid))
			addition := GetChildrenRecursively(n, hasToBeFree)
			for _, val := range addition {
				returnSlice = append(returnSlice, val)
			}
		}
	}
	return returnSlice

}

func FindFittingChild(currentNode *node, point *vector3, depth int) *node {
	if len(currentNode.uid) == depth {
		return currentNode
	}

	if currentNode.children[0] == nil {
		return currentNode
	}

	for _, child := range currentNode.children {
		if child.PointFits(point) {
			return FindFittingChild(child, point, depth)
		}
	}

	fmt.Println("----------------------------- DID NOT FIND A FITTING CHILD -----------------------------------")
	return nil
}

func GetChildrenRecursively(currentNode *node, hasToBeFree bool) []string {
	var returnSlice []string
	GetChildrenRecursivelyTask(currentNode, &returnSlice, hasToBeFree)
	return returnSlice
}

func GetChildrenRecursivelyTask(currentNode *node, returnSlice *[]string, hasToBeFree bool) {

	if currentNode.children[0] != nil {
		for i := 0; i < 8; i++ {
			GetChildrenRecursivelyTask(currentNode.children[i], returnSlice, hasToBeFree)
		}
	} else {
		if hasToBeFree {
			if currentNode.point != nil {
				return
			}
		}

		*returnSlice = append(*returnSlice, currentNode.uid)
	}
}

// GetNodeWithUid
func (tree *octree) GetNodeWithUid(uid string) *node {
	currentNode := tree.root
	for true {
		if currentNode.uid == uid {
			return currentNode
		}
		if len(currentNode.uid) < len(uid) {
			for _, child := range currentNode.children {
				if child.uid == uid {
					currentNode = child
					break
				}
			}
		} else {
			fmt.Println("----------------------------- DID NOT FIND A FITTING NODE -----------------------------------")
			return nil
		}
	}
	return nil
}

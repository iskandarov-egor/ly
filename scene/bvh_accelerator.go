package scene

import (
	"ly/geo"
	"ly/debug"
	"ly/util/math32"
	"sort"
	"fmt"
)

type Aggregate interface {
	RayIntersection(geo.Ray) (*ShapeHitPoint)
}

// implements Aggregate interface
type BVHNode struct {
	Shapes []Shape
	BoundingBox geo.Box
	Left   *BVHNode
	Right  *BVHNode
}

type BVHBuildNode struct {
	ShapeInfos []ShapeInfo
	BoundingBox geo.Box
	Left   *BVHBuildNode
	Right  *BVHBuildNode
	SplitAxis geo.Axis
}

type ShapeInfo struct {
	Shape Shape
	Box   geo.Box
	Center geo.Vec3
}

func Shapes2ShapeInfos(shapes []Shape) []ShapeInfo {
	ret := make([]ShapeInfo, len(shapes))
	for i, shape := range shapes {
		ret[i] = ShapeInfo{
			Shape: shape,
			Box: shape.BoundingBox(),
		}
		ret[i].Center = ret[i].Box.Min.Add(ret[i].Box.Max).Mul(0.5)
	}
	return ret
}

func ShapeInfos2Shapes(shapeInfos []ShapeInfo) []Shape {
	ret := make([]Shape, len(shapeInfos))
	for i, shapeInfo := range shapeInfos {
		ret[i] = shapeInfo.Shape
	}
	return ret
}

func StripBuildInfo(node *BVHBuildNode) *BVHNode {
	ret := BVHNode{
		Shapes: ShapeInfos2Shapes(node.ShapeInfos),
		BoundingBox: node.BoundingBox,
	}
	if node.Left != nil {
		ret.Left = StripBuildInfo(node.Left)
		ret.Right = StripBuildInfo(node.Right)
	}
	return &ret
}

func MakeBVH(shapes []Shape) *BVHNode {
	if len(shapes) == 0 {
		panic("no shapes")
	}
	shapeInfos := Shapes2ShapeInfos(shapes)
	root := BuildBVHNode(shapeInfos, 1, 100)
	ret := StripBuildInfo(root)
	return ret
}

var level = 0

func (n *BVHNode) RayIntersection(ray geo.Ray) (hp *ShapeHitPoint) {
	if !n.BoundingBox.Intersect(ray) {
		return nil
	}
	if n.Left == nil {
		//leaf, search the shapes
		return RayIntersectShapes(n.Shapes, ray)
	}
	hp1 := n.Left.RayIntersection(ray)
	hp2 := n.Right.RayIntersection(ray)
	if hp1 == nil {
		return hp2
	} else if hp2 == nil {
		return hp1
	}
	if hp1.RayT < hp2.RayT {
		return hp1
	} else {
		return hp2
	}
}

/*
// for debug
func (n *BVHNode) FindShape(s Shape) string {
	var walk func(node *BVHNode, path string) string
	walk = func(node *BVHNode, path string) string {
		if node.Left != nil {
			l := walk(node.Left, path + "L")
			r := walk(node.Right, path + "R")
			if l != "" {
				return l
			}
			if r != "" {
				return r
			}
		} else {
			for _, obj := range node.Objects {
				if obj.Shape == s {
					return path
				}
			}
		}
		return ""
	}
	return walk(n, "")
}
*/

func TestFindBVH(node *BVHNode, ray geo.Ray) int {
	isHit := node.BoundingBox.Intersect(ray)
	mx, my := -64, -375
	if !isHit {
		if debug.IX == mx && debug.IY == my {
			fmt.Println("nohit")
		}
		return -1
	} else {
		if node.Left == nil {
			if debug.IX == mx && debug.IY == my {
				fmt.Println("tupik", ray)
			}
			return 1
		}
		if debug.IX == mx && debug.IY == my {
			fmt.Println("left")
		}
		r1 := TestFindBVH(node.Left, ray)
		if debug.IX == mx && debug.IY == my {
			fmt.Println("right")
		}
		r2 := TestFindBVH(node.Right, ray)
		if debug.IX == mx && debug.IY == my {
			fmt.Println("r1r2", r1, r2)
		}
		if r1 != -1 {
			return 2*r1
		} else if r2 != -1 {
			return 2*r2+1
		} else {
			return -1
		}
	}
}
/*
If you are insisting on an in-place approach instead of the trivial standard return [arr.filter(predicate), arr.filter(notPredicate)] approach, that can be easily and efficiently achieved using two indices, running from both sides of the array and swapping where necessary:

function partitionInplace(arr, predicate) {
    var i=0, j=arr.length;
    while (i<j) {
        while (predicate(arr[i]) && ++i<j);
        if (i==j) break;
        while (i<--j && !predicate(arr[j]));
        if (i==j) break;
        [arr[i], arr[j]] = [arr[j], arr[i]];
        i++;
    }
    return i; // the index of the first element not to fulfil the predicate
}
*/

func partitionInplace(arr []ShapeInfo, axis geo.Axis, mid float32) int {
    i := 0
	j := len(arr)
    for i < j {
        for arr[i].Center.Axis(axis) <= mid {
			i++
			if i >= j {
				break
			}
		}
        if i == j {
			break
		}
		j--
        for (i < j) && (arr[j].Center.Axis(axis) > mid) {
			j--
		}
        if i == j {
			break
		}
        arr[i], arr[j] = arr[j], arr[i]
        i++
    }
    return i // the index of the first element not to fulfill the predicate
}

func BuildBVHNode(shapes []ShapeInfo, depth, maxDepth int) *BVHBuildNode {
	node := BVHBuildNode{}
	if len(shapes) == 1 || depth >= maxDepth {
		node.BoundingBox = geo.NewBox()
		for _, shape := range shapes {
			node.BoundingBox = node.BoundingBox.Union(shape.Shape.BoundingBox())
		}
		node.ShapeInfos = shapes
		return &node
	}
	var centerBounds geo.Box = geo.NewBox()
	for _, shape := range shapes {
		centerBounds.Include(shape.Center)
	}
	
	diag := centerBounds.Diagonal()
	maxExtent := math32.Max3(diag.X, diag.Y, diag.Z)
	if maxExtent == diag.X {
		node.SplitAxis = geo.AxisX
	} else if maxExtent == diag.Y {
		node.SplitAxis = geo.AxisY
	} else {
		node.SplitAxis = geo.AxisZ
	}

	if centerBounds.Min.Axis(node.SplitAxis) == centerBounds.Max.Axis(node.SplitAxis) {
		node.BoundingBox = geo.NewBox()
		for _, shape := range shapes {
			node.BoundingBox = node.BoundingBox.Union(shape.Shape.BoundingBox())
		}
		node.ShapeInfos = shapes
		return &node
	}

	midPoint := centerBounds.Min.Add(centerBounds.Max).Mul(0.5)
	iMiddle := partitionInplace(shapes, node.SplitAxis, midPoint.Axis(node.SplitAxis))

	//fmt.Println("midel", iMiddle)
	if iMiddle == 0 {
		// all triangles must be equal..
		node.BoundingBox = geo.NewBox()
		for _, shape := range shapes {
			node.BoundingBox = node.BoundingBox.Union(shape.Shape.BoundingBox())
		}
		node.ShapeInfos = shapes
		return &node
	}
	//for i := 0; i < iMiddle; i++ {
		//fmt.Println("LEFT", objects[i].Object.Shape)
	//}
	//for i := iMiddle; i < len(shapes); i++ {
		//fmt.Println("RGHT", objects[i].Object.Shape)
	//}
	//fmt.Println("axis", node.SplitAxis, "imid", iMiddle, "m", midPoint, "c", objects[iMiddle].Center)
	node.Left = BuildBVHNode(shapes[0:iMiddle], depth + 1, maxDepth)
	node.Right = BuildBVHNode(shapes[iMiddle:], depth + 1, maxDepth)
	//node.ShapeInfos = shapes
	node.BoundingBox = node.Left.BoundingBox.Union(node.Right.BoundingBox)
	return &node
}

func init() {
	_ = sort.Slice
}

package dynamictrie

import (
	"encoding/binary"
	"net"

	"github.com/junzexu/nuts/logging"
)

var logger = logging.GetLogger("dynamictrie")

type treeNode struct {
	Net      *V6Net
	Data     interface{}
	ChilBits uint64
	Chils    []*treeNode
}

// Tree .
type Tree struct {
	root *treeNode
}

// Add .
func (t *Tree) Add(net *V6Net, data interface{}) {
	t.root = add(t.root, &treeNode{Net: net, Data: data, ChilBits: 1, Chils: []*treeNode{nil, nil}})
}

func add(root, new *treeNode) *treeNode {
	if root == nil {
		// logger.Info("add %+v/%d  %+v, done", new.Net.ip, new.Net.maskBits-128+32, new.Data)
		return new
	}

	// logger.Info("add %+v/%d to %+v", root.Net.ip, root.Net.maskBits-128+32, new.Data)

	rel, ones := root.Net.CheckSub(new.Net)
	newRoot := root
	switch rel {
	case v6NetEqual:
		// logger.Info("equal")
		root.Data = new.Data
	case v6NetContainLeft:
		// logger.Info("new to left")
		root.Chils[0] = add(root.Chils[0], new)
	case v6NetContainRight:
		// logger.Info("new to right")
		root.Chils[1] = add(root.Chils[1], new)
	case v6NetContainedLeft:
		// logger.Info("root to left")
		newRoot = new
		newRoot.Chils[0] = root
	case v6NetContainedRight:
		// logger.Info("root to right")
		newRoot = new
		newRoot.Chils[1] = root
	case v6OvelapLeft:
		// logger.Info("split, new to right")
		newRoot = &treeNode{Net: &V6Net{root.Net.ip, ones}, Data: nil, ChilBits: 1, Chils: []*treeNode{root, new}}
	case v6OvelapRight:
		// logger.Info("split, new to left")
		newRoot = &treeNode{Net: &V6Net{root.Net.ip, ones}, Data: nil, ChilBits: 1, Chils: []*treeNode{new, root}}
	}

	return newRoot
}

// Get .
func (t *Tree) Get(ipStr string) (*V6Net, interface{}) {
	root := t.root
	var result interface{}
	var v6Net *V6Net

	ip := net.ParseIP(ipStr)
	ip = ip.To16()
	target := &V6Net{ip: ip, maskBits: 128}
	loopCnt := 0
	for root != nil {
		loopCnt++
		// logger.Info("matching %+v/%d", root.Net.ip, root.Net.maskBits-128+32)
		_, ones := root.Net.CheckSub(target)
		if ones >= root.Net.maskBits { // contains
			if root.Data != nil {
				result = root.Data
				v6Net = root.Net
			}

			// calculate child idx
			bByte := ones / 8
			eByte := (ones + root.ChilBits + 7) / 8
			offset := ones % 8
			data := make([]byte, 8)
			copy(data, ip[bByte:eByte])
			idx := binary.BigEndian.Uint64(data)
			childIdx := GetBits(idx, offset, root.ChilBits)

			// iterate for child
			root = root.Chils[childIdx]
		} else {
			break
		}
	}
	logger.Debug("loop cnt: %+v", loopCnt)
	return v6Net, result
}

// Compress .
func (t *Tree) Compress() {
	// BFS
	compress(t.root)
}

func compress(root *treeNode) {
	x := bfs(root)
	levels := uint64(len(x) - 1)
	childs := x[levels]
	if levels > 1 && 0x01<<levels == len(childs) {
		logger.Info("with level %+v, %0x", levels, len(childs))

		root.ChilBits = uint64(levels)
		root.Chils = childs
	}

	for _, child := range root.Chils {
		if child != nil {
			compress(child)
		}
	}
}

func bfs(root *treeNode) [][]*treeNode {
	result := [][]*treeNode{}
	prev := []*treeNode{root}

	shouldBreak := false

	for len(prev) > 0 && !shouldBreak {
		// logger.Info("bfs level %+v, %0x", len(result), prev)
		result = append(result, prev)
		nilCnt := 0
		next := []*treeNode{}
		for _, x := range prev {
			if x == nil {
				nilCnt++
				next = append(next, nil, nil)
			} else if len(x.Chils) == 2 {
				for _, child := range x.Chils {
					if child != nil {
						if child.Data != nil {
							shouldBreak = true
							break
						}
					}
					next = append(next, child)
				}
			} else {
				panic("only support binary tree")
			}
		}

		if float64(nilCnt) >= float64(len(prev))*0.5 {
			shouldBreak = true
		}

		prev = next
	}

	return result
}

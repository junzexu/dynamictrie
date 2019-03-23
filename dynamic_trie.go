package dynamictrie

import (
	"net"

	"github.com/junzexu/nuts/logging"
)

var logger = logging.GetLogger("dynamictrie")

func ipbit(n net.IP, bIndex uint) byte {
	bs, bit := bIndex/8, bIndex%8
	mask := byte(1) << (uint8(7 - bit))
	return (n[bs] & mask)
}

// TreeNode .
type TreeNode struct {
	Prefix     net.IP
	PrefixBits uint
	CHilds     [2]*TreeNode
	Data       interface{} // where ChildBits=0, its a leaf node
}

// Tree .
type Tree struct {
	root *TreeNode
}

// Add .
func (tr *Tree) Add(n net.IP, p uint, data interface{}) error {
	if tr.root == nil {
		tr.root = &TreeNode{
			Prefix:     n,
			PrefixBits: p,
			Data:       data,
		}
		// logger.Info("add root node: %+v", tr.root)
		return nil
	}

	root := tr.root
	i := uint(0)
	for {
		fullMatch := true
		for ; i < root.PrefixBits && i < p; i++ {
			if ipbit(root.Prefix, i) != ipbit(n, i) {
				fullMatch = false
				break
			}
		}
		// logger.Info("common leading 1 size: %d, fullMatch=%+v, %02x, %02x", i, fullMatch, []byte(root.Prefix), []byte(n))
		if fullMatch {
			if root.PrefixBits > p {
				tmpNode := *root // copy
				root.Data = data
				root.PrefixBits = i

				if ipbit(tmpNode.Prefix, i) > 0 {
					// logger.Info("full match, as parent node: %+v, origin as right", root)
					root.CHilds = [2]*TreeNode{nil, &tmpNode}
				} else {
					// logger.Info("full match, as parent node: %+v, origin as left", root)
					root.CHilds = [2]*TreeNode{&tmpNode, nil}
				}

			} else if root.PrefixBits == p {
				// logger.Info("full match, update data")
				root.Data = data
			} else {
				isSet := 0
				if ipbit(n, i) > 0 {
					isSet = 1
				}
				if root.CHilds[isSet] != nil {
					root = root.CHilds[isSet]
					// logger.Info("full match, for next %+v, isLeft=%d", root, isSet)
					continue
				}
				root.CHilds[isSet] = &TreeNode{
					Data:       data,
					Prefix:     n,
					PrefixBits: p,
				}
				// logger.Info("full match, create child node, isLeft=%d", root.CHilds[isSet], isSet)
			}
		} else {
			tmpNode := *root // copy
			newNode := TreeNode{
				Data:       data,
				Prefix:     n,
				PrefixBits: p,
			}
			root.Data = nil
			root.Prefix = n
			root.PrefixBits = i
			if ipbit(tmpNode.Prefix, i) > 0 {
				root.CHilds = [2]*TreeNode{&newNode, &tmpNode}
			} else {
				root.CHilds = [2]*TreeNode{&tmpNode, &newNode}
			}
			// logger.Info("not full match split: root=%+v, left=%+v, right=%+v", root, root.CHilds[0], root.CHilds[1])
		}
		break
	}
	return nil
}

// FindLCF .
func (tr *Tree) FindLCF(n net.IP) interface{} {
	var result interface{}
	root := tr.root
	i := uint(0)
	for {
		if root == nil {
			return result
		}

		for ; i < root.PrefixBits; i++ {
			if ipbit(root.Prefix, i) != ipbit(n, i) {
				logger.Info("not match: %+v, %+v, %d", n, root, i)
				return result
			}
		}
		// match
		result = root.Data

		if ipbit(n, i) > 0 {
			root = root.CHilds[1]
			logger.Info("visit right: %+v, %+v", n, root)
		} else {
			root = root.CHilds[0]
			logger.Info("visit left: %+v, %+v", n, root)
		}
	}
}

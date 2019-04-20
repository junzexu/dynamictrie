package dynamictrie

import (
	"encoding/binary"
	"math"
	"math/bits"
	"net"
)

// V6Net .
type V6Net struct {
	ip       net.IP // IPv6
	maskBits uint64
}

// NewV6Net .
func NewV6Net(str string) *V6Net {
	ip, net, _ := net.ParseCIDR(str)
	ones, bits := net.Mask.Size()

	if bits == 32 {
		ip = ip.To16()
		ones = ones + (128 - 32)
	}

	rst := &V6Net{
		ip:       ip,
		maskBits: uint64(ones),
	}

	// ogger.Info("%+v, %+v", str, rst)
	return rst
}

// V6NetRel .
type V6NetRel int

const (
	// a.CheckSub(b)
	v6NetEqual          V6NetRel = 0 // a == b
	v6NetContainLeft    V6NetRel = 1 // b 作为 a 左子节点
	v6NetContainRight   V6NetRel = 2 // b 作为 a 右子节点
	v6NetContainedLeft  V6NetRel = 3 // a 作为 b 左子节点
	v6NetContainedRight V6NetRel = 4 // a 作为 b 右子节点
	v6OvelapLeft        V6NetRel = 5 // a 左 b 右
	v6OvelapRight       V6NetRel = 6 // a 右 b 左
)

// CheckSub @return
// V6NetRel 表示相对的，被检查参数应该在树种的相对位置
func (n *V6Net) CheckSub(t *V6Net) (V6NetRel, uint64) {
	// check high 64bit
	l := binary.BigEndian.Uint64(n.ip[:8])
	r := binary.BigEndian.Uint64(t.ip[:8])
	var lbits, rbits uint64 = 64, 64
	if n.maskBits < lbits {
		lbits = n.maskBits
	}
	if t.maskBits < rbits {
		rbits = t.maskBits
	}

	result, commonOnes := checkSub(l, r, lbits, rbits)
	if result != v6NetEqual {
		return result, commonOnes
	}

	// check low 64 bit
	l = binary.BigEndian.Uint64(n.ip[8:16])
	r = binary.BigEndian.Uint64(t.ip[8:16])
	lbits, rbits = 0, 0
	if n.maskBits > 64 {
		lbits = n.maskBits - 64
	}
	if t.maskBits > 64 {
		rbits = t.maskBits - 64
	}
	result, commonOnes = checkSub(l, r, lbits, rbits)
	commonOnes += 64
	return result, commonOnes
}

func checkSub(lip, rip uint64, lMaskOnes, rMaskOnes uint64) (V6NetRel, uint64) {
	l := lip & (math.MaxUint64 << (64 - lMaskOnes))
	r := rip & (math.MaxUint64 << (64 - rMaskOnes))

	// logger.Debug("%0x/%d, %0x/%d", l, lMaskOnes, r, rMaskOnes)
	if lMaskOnes != rMaskOnes || l != r {
		commonOnes := uint64(bits.LeadingZeros64((l ^ r)))
		if commonOnes > lMaskOnes {
			commonOnes = lMaskOnes
		}
		if commonOnes > rMaskOnes {
			commonOnes = rMaskOnes
		}
		// logger.Debug("%0x/%d, %0x/%d, common %d", l, lMaskOnes, r, rMaskOnes, commonOnes)
		if commonOnes == lMaskOnes {
			// contain
			if GetBits(r, commonOnes, 1) == 0 {
				return v6NetContainLeft, commonOnes
			}
			return v6NetContainRight, commonOnes
		}

		nextBit := GetBits(l, commonOnes, 1)
		if commonOnes == rMaskOnes {
			// contained
			if nextBit == 0 {
				return v6NetContainedLeft, commonOnes
			}
			return v6NetContainedRight, commonOnes
		}

		// overlap
		if nextBit == 0 {
			return v6OvelapLeft, commonOnes
		}
		return v6OvelapRight, commonOnes
	}

	return v6NetEqual, lMaskOnes
}

// GetBits .
func GetBits(x uint64, p uint64, size uint64) uint64 {
	return ((x << p) >> (64 - size))
}

package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"math/bits"
	"net"
	"os"
	"strings"

	"github.com/junzexu/nuts/logging"

	dynamictrie "github.com/junzexu/dynamicTrie"
)

var logger = logging.GetLogger("main")

func main() {
	csvFobj, _ := os.Open("ad_ip_info.csv")
	csvReader := csv.NewReader(csvFobj)
	csvReader.Comma = '\t'

	tree := &dynamictrie.Tree{}
	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		begin := net.ParseIP(line[0]).To4()
		end := net.ParseIP(line[1]).To4()
		maskBits := ip4GetMaskBits(begin, end)

		logger.Info("[%+v, %+v], mask=%+v", begin, end, maskBits)
		tree.Add(begin.To4(), uint(maskBits), strings.Join(line, "\t"))
	}

	scan := bufio.NewReader(os.Stdin)
	for {
		line, _ := scan.ReadBytes('\n')
		ipStr := strings.TrimSpace(string(line))
		if len(ipStr) > 0 {
			ip := net.ParseIP(ipStr).To4()
			rst := tree.FindLCF(ip)
			logger.Info("%+v => %+v", ip, rst)
		}
	}
}

func ip4GetMaskBits(b, e net.IP) int {
	size := ipv4AsUint32(e) - ipv4AsUint32(b)
	return bits.LeadingZeros32(size)
}

func ipv4AsUint32(ip net.IP) uint32 {
	result := uint32(0)
	for _, b := range ip {
		result = ((result << 8) | uint32(b))
	}
	return result
}

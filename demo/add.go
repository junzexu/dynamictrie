package main

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"io"
	"math"
	"math/bits"
	"net"
	"os"
	"strings"

	"net/http"
	_ "net/http/pprof"

	"github.com/junzexu/nuts/logging"

	dynamictrie "github.com/junzexu/dynamicTrie"
)

var logger = logging.GetLogger("main")

func main() {
	go func() {
		logger.Warn("running http")
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	logger.Info("begin")
	// logging.DefaultConsoleHandler.SetLevel(logging.WarnningLevel)
	csvFobj, _ := os.Open("ad_ip_info.csv")
	csvReader := csv.NewReader(csvFobj)
	csvReader.Comma = '\t'

	tree := &dynamictrie.Tree{}
	cnt := 0
	for {
		cnt++
		//logger.Warn("loop, %+v", cnt)
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		begin := net.ParseIP(line[0]).To4()
		end := net.ParseIP(line[1]).To4()
		maskBits := ip4GetMaskBits(begin, end)
		mask := make([]byte, 4)
		binary.BigEndian.PutUint32(mask, math.MaxUint32<<uint8(32-maskBits))
		ipNet := net.IPNet{
			IP:   begin,
			Mask: net.IPMask(mask),
		}

		// ger.Info("[%+v, %+v], %+v, mask=%+v", begin, end, maskBits, ipNet)
		tree.Add(dynamictrie.NewV6Net(ipNet.String()), strings.Join(line, "\t"))

		if cnt > 3 {
			// break
		}
	}
	logger.Warn("loading done")

	tree.Compress()

	// dynamictrie.Compress(tree, 0.2)
	// logger.Info("compress done")

	scan := bufio.NewReader(os.Stdin)
	for {
		line, _ := scan.ReadBytes('\n')
		ipStr := strings.TrimSpace(string(line))
		if len(ipStr) > 0 {
			_, rst := tree.Get(ipStr)
			logger.Info("%+v => %+v", ipStr, rst)
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

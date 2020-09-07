package glog

import (
	"net"
	"runtime"
	"strconv"
	"strings"
)

// getFrame 获取调用函数信息
func getFrame(skipFrames int) *runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := &runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				*frame = frameCandidate
			}
		}
	}

	return frame
}

// getFuncName 解析函数名
func getFuncName(function string) string {
	idx := strings.LastIndexByte(function, '.')
	if idx > 0 {
		return function[idx+1:]
	}

	return function
}

func toUpper(x byte) byte {
	if x <= 'Z' {
		return x
	}

	x -= 'a' - 'A'
	return x
}

func toString(num int) string {
	return strconv.Itoa(num)
}

func isDigit(x byte) bool {
	return x >= '0' && x <= '9'
}

func toDigit(x byte) int {
	return int(x - '0')
}

// getLocalIp 获取本地IP
func getLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// +build darwin dragonfly freebsd netbsd openbsd

package glog

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA

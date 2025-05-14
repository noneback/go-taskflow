package utils

import (
	"os"
	"runtime/pprof"
)

type ProfileType int

type PprofUtils struct {
	f       *os.File
	profile ProfileType
}

func NewPprofUtils(profile ProfileType, output string) *PprofUtils {
	p := &PprofUtils{}
	f, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	p.f = f
	p.profile = profile

	return p
}

const (
	CPU ProfileType = iota
	HEAP
)

func (p *PprofUtils) StartProfile() {
	switch p.profile {
	case CPU:
		if err := pprof.StartCPUProfile(p.f); err != nil {
			panic(err)
		}
	case HEAP:
		if err := pprof.WriteHeapProfile(p.f); err != nil {
			panic(err)
		}
	default:
		panic("unsupported profile type")
	}

}

func (p *PprofUtils) StopProfile() {
	defer p.f.Close()

	switch p.profile {
	case CPU:
		pprof.StopCPUProfile()
	case HEAP:
	default:
		panic("unsupported profile type")
	}
}

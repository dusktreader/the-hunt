package main

import "runtime/debug"

func Version() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		return info.Main.Version
	}
	return ""
}

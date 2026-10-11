//go:build with_boxdd && (!windows && !(linux && !android))

package main

func newPlatformInterface(daemon *Daemon) (daemonPlatform, error) {
	return nil, nil
}

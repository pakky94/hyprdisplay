package backend

import (
	"fmt"
	"net"
	"os"
	"slices"
)

type HyprCtl struct {
	cmdSocket   net.Conn
	eventSocket net.Conn
}

func (h *HyprCtl) Close() {
	h.cmdSocket.Close()
	h.eventSocket.Close()
}

func (h *HyprCtl) SendRaw(cmd []byte) error {
	_, err := h.cmdSocket.Write(cmd)
	return err
}

func OpenConn() (*HyprCtl, error) {
	xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR")
	his := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")

	cmdSocketPath := fmt.Sprintf("%s/hypr/%s/.socket.sock", xdgRuntimeDir, his)
	eventSocketPath := fmt.Sprintf("%s/hypr/%s/.socket2.sock", xdgRuntimeDir, his)

	cmdSocket, err := net.Dial("unix", cmdSocketPath)

	if err != nil {
		return nil, fmt.Errorf("Error opening %s %w", cmdSocketPath, err)
	}

	eventSocket, err := net.Dial("unix", eventSocketPath)

	if err != nil {
		cmdSocket.Close()
		return nil, fmt.Errorf("Error opening %s %w", eventSocketPath, err)
	}

	return &HyprCtl{
		cmdSocket,
		eventSocket,
	}, nil
}

func (hc *HyprCtl) Loop(ret chan string) error {
	readBuf := make([]byte, 1024)
	cmdBuf := make([]byte, 0, 4096)

	for true {
		n, err := hc.eventSocket.Read(readBuf)

		if n > 0 {
			cmdBuf = append(cmdBuf, readBuf[:n]...)
			ind := slices.Index(cmdBuf, '\n')

			for ind >= 0 {
				cmd := string(cmdBuf[:ind])
				cmdBuf = cmdBuf[ind+1:]
				ret <- cmd
				ind = slices.Index(cmdBuf, '\n')
			}
		}

		if e, ok := err.(interface{ Timeout() bool }); ok && e.Timeout() {
			// handle timeout
		} else if err != nil {
			return err
		}
	}

	return nil
}

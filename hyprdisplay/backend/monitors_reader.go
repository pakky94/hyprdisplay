package backend

import (
	"encoding/json"
	"os/exec"
)

type MonitorStatus struct {
	Id          int
	Name        string
	Description string
	Disabled    bool
}

func ReadHyprMonitors() ([]MonitorStatus, error) {
	cmd := exec.Command("hyprctl", "monitors", "all", "-j")
	res, err := cmd.Output()

	if err != nil {
		return nil, err
	}

	var monitors []MonitorStatus

	err = json.Unmarshal(res, &monitors)
	if err != nil {
		return nil, err
	}

	return monitors, nil
}

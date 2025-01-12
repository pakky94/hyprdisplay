package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

type MonitorStatus struct {
	Id          int
	Name        string
	Description string
	Disabled    int
	Width       int
	Height      int
	RefreshRate string
	X           int
	Y           int
	Scale       string
	Transform   int
}

type msInternal struct {
	Id          int
	Name        string
	Description string
	Disabled    bool
	Width       int
	Height      int
	RefreshRate json.Number
	X           int
	Y           int
	Scale       json.Number
	Transform   int
}

func ToKey(ms []MonitorStatus) string {
	slices.SortFunc(ms, func(a MonitorStatus, b MonitorStatus) int {
		return strings.Compare(a.Description, b.Description)
	})

	descriptions := make([]string, 0, len(ms))

	for _, m := range ms {
		descriptions = append(descriptions, m.Description)
	}

	return strings.Join(descriptions, ";")
}

func ReadHyprMonitors() ([]MonitorStatus, error) {
	cmd := exec.Command("hyprctl", "monitors", "all", "-j")
	res, err := cmd.Output()

	stdErr := ""
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		stdErr = string(exitErr.Stderr)
	}

	if err != nil {
		return nil, fmt.Errorf("Error reading monitors %w, out: %q, details: %q", err, string(res), stdErr)
	}

	var monitors []msInternal

	err = json.Unmarshal(res, &monitors)
	if err != nil {
		return nil, err
	}

	ms := make([]MonitorStatus, 0, len(monitors))
	for _, m := range monitors {
		mapped := MonitorStatus{
			Id:          m.Id,
			Name:        m.Name,
			Description: m.Description,
			Width:       m.Width,
			Height:      m.Height,
			RefreshRate: string(m.RefreshRate),
			X:           m.X,
			Y:           m.Y,
			Scale:       string(m.Scale),
			Transform:   m.Transform,
		}

		if m.Disabled {
			mapped.Disabled = 1
		}

		ms = append(ms, mapped)
	}

	return ms, nil
}

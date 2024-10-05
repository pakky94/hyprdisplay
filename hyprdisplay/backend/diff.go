package backend

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

func Apply(cmds []string) error {
	if len(cmds) == 0 {
		return nil
	}

	join := strings.Join(cmds, " ; ")

	return exec.Command("hyprctl", "--batch", join).Run()
}

func Diff(actual []MonitorStatus, target []MonitorStatus) []string {
	res := make([]string, 0)

	for _, a := range actual {
		i := slices.IndexFunc(target, func(m MonitorStatus) bool {
			return m.Description == a.Description
		})

		if i == -1 {
			continue
		}

		cmd := singleDiff(a, target[i])

		if cmd != "" {
			res = append(res, cmd)
		}
	}

	return res
}

func singleDiff(actual MonitorStatus, target MonitorStatus) string {
	if actual.Disabled == 0 && target.Disabled == 1 {
		return fmt.Sprintf("keyword monitor %s, disable", target.Name)
	}

	if actual.Disabled != target.Disabled ||
		actual.Width != target.Width ||
		actual.Height != target.Height ||
		actual.RefreshRate != target.RefreshRate ||
		actual.X != target.X ||
		actual.Y != target.Y ||
		actual.Scale != target.Scale ||
		actual.Transform != target.Transform {

		t := fmt.Sprintf(
			"keyword monitor %s, %dx%d@%s, %dx%d, %s",
			target.Name,
			target.Width,
			target.Height,
			target.RefreshRate,
			target.X,
			target.Y,
			target.Scale,
		)

		if target.Transform != 0 {
			return t + ", transform, " + fmt.Sprint(target.Transform)
		} else {
			return t
		}
	}

	return ""
}

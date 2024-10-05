package backend

import (
	"fmt"
	"strings"
)

func Daemonize() {
	dbPath, err := DefaultDbPath()
	if err != nil {
		panic(err)
	}

	db, err := InitDb(dbPath, DB_NAME)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	currentSetup, err := ReadHyprMonitors()
	if err != nil {
		panic(err)
	}

	dbSetup, err := FindSetup(db, ToKey(currentSetup))
	if err != nil {
		panic(err)
	}

	if len(dbSetup) == 0 {
		println("no setup found in db")

		err := SaveSetup(db, ToKey(currentSetup), currentSetup)
		if err != nil {
			panic(err)
		}
	} else {
		cmds := Diff(currentSetup, dbSetup)
		err := Apply(cmds)
		if err != nil {
			panic(err)
		}
	}

	ctl, err := OpenConn()
	if err != nil {
		panic(err)
	}
	defer ctl.Close()

	cmdChan := make(chan string)
	go RunDaemon(cmdChan)
	ctl.Loop(cmdChan)

	// ctl.SendRaw([]byte("/keyword monitor eDP-1, disable"))
	// ctl.SendRaw([]byte("/keyword monitor eDP-1,preferred,0x0,1,transform,3"))
}

func RunDaemon(cmdChan chan string) {
	for true {
		cmd := <-cmdChan
		if strings.HasPrefix(cmd, "monitoradded>>") || strings.HasPrefix(cmd, "monitorremoved>>") {
			println(cmd)
			monitors, err := ReadHyprMonitors()

			if err != nil {
				panic(err)
			}

			println(fmt.Sprintf("%+v", monitors))
		}
	}
}

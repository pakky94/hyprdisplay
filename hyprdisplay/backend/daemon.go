package backend

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

var Verbose bool = false

type Daemon struct {
	db  *sql.DB
	ctl *HyprCtl
}

func (d *Daemon) Close() {
	d.db.Close()
	d.ctl.Close()
}

func Daemonize() *Daemon {
	dbPath, err := DefaultDbPath()
	if err != nil {
		panic(err)
	}

	db, err := InitDb(dbPath, DB_NAME)
	if err != nil {
		panic(err)
	}

	ctl, err := OpenConn()
	if err != nil {
		panic(err)
	}

	err = tryApplySavedSetup(db)
	if err != nil {
		panic(err)
	}

	cmdChan := make(chan string)
	go RunDaemon(db, cmdChan)
	go ctl.Loop(cmdChan)

	// ctl.SendRaw([]byte("/keyword monitor eDP-1, disable"))
	// ctl.SendRaw([]byte("/keyword monitor eDP-1,preferred,0x0,1,transform,3"))
	return &Daemon{
		db,
		ctl,
	}
}

func tryApplySavedSetup(db *sql.DB) error {
	currentSetup, err := ReadHyprMonitors()
	if err != nil {
		panic(err)
	}

	dbSetup, err := FindSetup(db, ToKey(currentSetup))
	if err != nil {
		return err
	}

	if len(dbSetup) == 0 {
		if Verbose {
			log.Printf(fmt.Sprintf("Saving new setup for monitors %q", ToKey(currentSetup)))
		}

		err := SaveSetup(db, ToKey(currentSetup), currentSetup)
		if err != nil {
			return err
		}
	} else {
		if Verbose {
			log.Printf(fmt.Sprintf("Applying saved setup %q", ToKey(currentSetup)))
		}

		cmds := Diff(currentSetup, dbSetup)
		err := Apply(cmds, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func RunDaemon(db *sql.DB, cmdChan chan string) {
	for true {
		cmd := <-cmdChan

		// if Verbose {
		log.Printf(fmt.Sprintf("Recieved event from hyprland: %q", cmd))
		// }

		if strings.HasPrefix(cmd, "monitoradded>>") ||
			strings.HasPrefix(cmd, "monitorremoved>>") {
			err := tryApplySavedSetup(db)
			if err != nil {
				panic(err)
			}
		}
	}
}

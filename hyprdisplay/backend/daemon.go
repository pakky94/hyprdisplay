package backend

import (
	"database/sql"
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

	ctl, err := OpenConn()
	if err != nil {
		panic(err)
	}
	defer ctl.Close()

	err = tryApplySavedSetup(db)
	if err != nil {
		panic(err)
	}

	// cmdChan := make(chan string)
	// go RunDaemon(db, cmdChan)
	// ctl.Loop(cmdChan)

	// ctl.SendRaw([]byte("/keyword monitor eDP-1, disable"))
	// ctl.SendRaw([]byte("/keyword monitor eDP-1,preferred,0x0,1,transform,3"))
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
		println("a")
		err := SaveSetup(db, ToKey(currentSetup), currentSetup)
		if err != nil {
			return err
		}
	} else {
		println("b")
		cmds := Diff(currentSetup, dbSetup)
		err := Apply(cmds)
		if err != nil {
			return err
		}
	}

	return nil
}

func RunDaemon(db *sql.DB, cmdChan chan string) {
	for true {
		cmd := <-cmdChan
		if strings.HasPrefix(cmd, "monitoradded>>") ||
			strings.HasPrefix(cmd, "monitorremoved>>") {
			err := tryApplySavedSetup(db)
			if err != nil {
				panic(err)
			}

			// println(fmt.Sprintf("%+v", monitors))
		}
	}
}

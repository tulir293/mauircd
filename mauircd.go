// mauIRCd - The IRC bouncer/backend system for mauIRC clients.
// Copyright (C) 2016 Tulir Asokan

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
	"maunium.net/go/libmauirc"
	cfg "maunium.net/go/mauircd/config"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/web"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var nws = flag.StringP("config", "c", "/etc/mauircd/", "The path to mauIRCd configurations")
var config mauircdi.Configuration

func init() {
	libmauirc.Version = "mauIRC 0.1"
}

func main() {
	flag.Parse()

	config = cfg.NewConfig(*nws)
	err := config.Load()
	if err != nil {
		panic(err)
	}

	err = database.Load(config.GetSQLString())
	if err != nil {
		panic(err)
	}

	config.Connect()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nClosing mauIRCd")
		config.GetUsers().ForEach(func(user mauircdi.User) {
			user.GetNetworks().ForEach(func(net mauircdi.Network) {
				net.Close()
				net.SaveScripts(config.GetPath())
				net.Save()
			})
		})
		time.Sleep(2 * time.Second)
		database.Close()
		config.Save()
		os.Exit(0)
	}()
	web.Load(config)
}

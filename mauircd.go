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
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/irc"
	"maunium.net/go/mauircd/web"
)

func main() {
	flag.Parse()

	database.Load("root", flag.Arg(0), "127.0.0.1", 3306, "mauircd")
	irc.TmpNet = irc.Create("pvlnet", "mauircd", "mauircd", "mauircd@maunium.net", "", "irc.fixme.fi", 6697, true)
	web.Load("127.0.0.1", 29304)
}

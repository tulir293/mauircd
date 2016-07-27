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

// Package web contains the HTTP server
package web

import (
	"encoding/json"
	"io/ioutil"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/plugin"
	"net/http"
	"strings"
)

const (
	all    = "all"
	global = "global"
)

func script(w http.ResponseWriter, r *http.Request) {
	authd, user := checkAuth(w, r)
	if !authd {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	args := strings.Split(r.RequestURI, "/")[2:]
	switch r.Method {
	case http.MethodGet:
		getScripts(w, r, args, user)
	case http.MethodDelete:
		deleteScript(w, r, args, user)
	case http.MethodPut:
		putScript(w, r, args, user)
	case http.MethodPost:
		postScript(w, r, args, user)
	default:
		w.Header().Add("Allow", http.MethodGet+","+http.MethodDelete+","+http.MethodPut+","+http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func postScript(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newPath := string(data)
	parts := strings.Split(newPath, ",")
	if len(parts) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var scripts []mauircdi.Script
	var success bool
	if args[0] == global {
		scripts = user.GetGlobalScripts()
		success = user.RemoveGlobalScript(args[1])
	} else {
		net := user.GetNetwork(args[0])
		if net == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		scripts = net.GetScripts()
		success = net.RemoveScript(args[1])
	}

	if !success {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	success = false

	var script mauircdi.Script
	for _, s := range scripts {
		if s.GetName() == args[1] {
			script = plugin.Script{Name: parts[1], TheScript: s.GetScript()}
			success = true
			break
		}
	}
	if !success {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if parts[0] == global {
		user.AddGlobalScript(script)
	} else {
		net := user.GetNetwork(parts[0])
		if net == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		net.AddScript(script)
	}
}

func putScript(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	script := plugin.Script{Name: args[1], TheScript: string(data)}

	if args[0] == global {
		user.AddGlobalScript(script)
	} else {
		net := user.GetNetwork(args[0])
		if net == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		net.AddScript(script)
	}
}

func getScripts(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	var scripts []mauircdi.Script
	if args[0] == all {
		scripts = user.GetGlobalScripts()
		user.GetNetworks().ForEach(func(net mauircdi.Network) {
			scripts = append(scripts, net.GetScripts()...)
		})
	} else if args[0] == global {
		scripts = user.GetGlobalScripts()
	} else {
		net := user.GetNetwork(args[0])
		if net == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		scripts = net.GetScripts()
	}

	data, err := json.Marshal(scripts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func deleteScript(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	if args[0] == "global" {
		if !user.RemoveGlobalScript(args[1]) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		net := user.GetNetwork(args[0])
		if net == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if !net.RemoveScript(args[1]) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

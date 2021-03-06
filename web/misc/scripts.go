// mauIRC-server - The IRC bouncer/backend system for mauIRC clients.
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

// Package misc contains HTTP-only misc handlers
package misc

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"maunium.net/go/mauirc-server/common/errors"
	"maunium.net/go/mauirc-server/interfaces"
	"maunium.net/go/mauirc-server/plugin"
	"maunium.net/go/mauirc-server/web/auth"
)

const (
	all    = "all"
	global = "global"
)

// Script HTTP handler
func Script(w http.ResponseWriter, r *http.Request) {
	authd, user := auth.Check(w, r)
	if !authd {
		errors.Write(w, errors.NotAuthenticated)
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
		w.Header().Add("Allow", strings.Join([]string{http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodPost}, ","))
		errors.Write(w, errors.InvalidMethod)
	}
}

func postScript(w http.ResponseWriter, r *http.Request, args []string, user interfaces.User) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errors.Write(w, errors.BodyNotFound)
		return
	}
	newPath := string(data)
	parts := strings.Split(newPath, ",")
	if len(parts) != 2 {
		errors.Write(w, errors.InvalidBodyFormat)
		return
	}

	var scripts []interfaces.Script
	var success bool
	if args[0] == global {
		scripts = user.GetGlobalScripts()
		success = user.RemoveGlobalScript(args[1])
	} else {
		net := user.GetNetwork(args[0])
		if net == nil {
			errors.Write(w, errors.NetworkNotFound)
			return
		}
		scripts = net.GetScripts()
		success = net.RemoveScript(args[1])
	}

	if !success {
		errors.Write(w, errors.ScriptNotFound)
		return
	}

	var script interfaces.Script
	for _, s := range scripts {
		if s.GetName() == args[1] {
			script = plugin.Script{Name: parts[1], TheScript: s.GetScript()}
			break
		}
	}

	if parts[0] == global {
		user.AddGlobalScript(script)
	} else {
		net := user.GetNetwork(parts[0])
		if net == nil {
			errors.Write(w, errors.NetworkNotFound)
			return
		}
		net.AddScript(script)
	}
}

func putScript(w http.ResponseWriter, r *http.Request, args []string, user interfaces.User) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errors.Write(w, errors.BodyNotFound)
		return
	}
	script := plugin.Script{Name: args[1], TheScript: string(data)}

	if args[0] == global {
		user.AddGlobalScript(script)
	} else {
		net := user.GetNetwork(args[0])
		if net == nil {
			errors.Write(w, errors.NetworkNotFound)
			return
		}
		net.AddScript(script)
	}
}

func getScripts(w http.ResponseWriter, r *http.Request, args []string, user interfaces.User) {
	var scripts []interfaces.Script
	if args[0] == all {
		scripts = user.GetGlobalScripts()
		user.GetNetworks().ForEach(func(net interfaces.Network) {
			scripts = append(scripts, net.GetScripts()...)
		})
	} else if args[0] == global {
		scripts = user.GetGlobalScripts()
	} else {
		net := user.GetNetwork(args[0])
		if net == nil {
			errors.Write(w, errors.NetworkNotFound)
			return
		}
		scripts = net.GetScripts()
	}

	data, err := json.Marshal(scripts)
	if err != nil {
		errors.Write(w, errors.Internal)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func deleteScript(w http.ResponseWriter, r *http.Request, args []string, user interfaces.User) {
	if args[0] == "global" {
		if !user.RemoveGlobalScript(args[1]) {
			errors.Write(w, errors.ScriptNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		net := user.GetNetwork(args[0])
		if net == nil {
			errors.Write(w, errors.NetworkNotFound)
			return
		}
		if !net.RemoveScript(args[1]) {
			errors.Write(w, errors.ScriptNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

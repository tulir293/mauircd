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

// Package auth contains the authentication system
package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"maunium.net/go/mauirc-server/common/errors"
	"maunium.net/go/mauirc-server/web/util"
)

// Register HTTP handler
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		errors.Write(w, errors.InvalidMethod)
		return
	}

	dec := json.NewDecoder(r.Body)
	var af authform
	err := dec.Decode(&af)

	if err != nil || len(af.Email) == 0 || len(af.Password) == 0 {
		errors.Write(w, errors.MissingFields)
		return
	}

	user, token, timeStamp := config.CreateUser(af.Email, af.Password)
	if user == nil {
		log.Debugf("%s tried to register existing user %s\n", util.GetIP(r), af.Email)
		errors.Write(w, errors.EmailUsed)
		return
	}
	log.Debugf("%s registered %s\n", util.GetIP(r), af.Email)

	if config.GetMail().IsEnabled() {
		config.GetMail().Send(user.GetEmail(), "mauIRC account created", "account-created", map[string]interface{}{
			"SenderIP":   util.GetIP(r),
			"ServerAddr": config.GetExternalAddr(),
			"ServerIP":   config.GetAddr(),
			"Token":      token,
			"Expiry":     timeStamp.Format(time.RFC1123),
		})
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

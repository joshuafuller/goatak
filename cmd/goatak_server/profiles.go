package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kdudkov/goatak/cmd/goatak_server/mp"
	"github.com/kdudkov/goatak/pkg/model"
)

func NewUserPrefsFile(prefix string, user *model.Device) *mp.PrefFile {
	conf := mp.NewUserProfilePrefFile(prefix)
	if user.Callsign != "" {
		conf.AddParam("locationCallsign", user.Callsign)
	}

	if user.Team != "" {
		conf.AddParam("locationTeam", user.Team)
	}

	if user.Role != "" {
		conf.AddParam("atakRoleType", user.Role)
	}

	if user.CotType != "" {
		conf.AddParam("locationUnitType", user.CotType)
	}

	return conf
}

func (app *App) GetProfileFiles(username, uid string) []mp.FileContent {
	res := make([]mp.FileContent, 0)
	prefix := fmt.Sprintf("%x", md5.Sum([]byte(username)))

	if app.users != nil && username != "" {
		if userInfo := app.users.Get(username); userInfo != nil {
			if userInfo.HasProfile() {
				app.logger.Debug("add user prefs")

				f := NewUserPrefsFile(prefix, userInfo)
				res = append(res, f)
			}
		}
	}

	if f, err := mp.NewFsFile(prefix+"/defaults.pref", filepath.Join(app.config.DataDir(), "defaults.pref")); err == nil {
		app.logger.Debug("add default.prefs")

		res = append(res, f)
	}

	if paths, err := os.ReadDir(filepath.Join(app.config.DataDir(), "maps")); err == nil {
		for _, p := range paths {
			if !p.IsDir() && strings.HasSuffix(p.Name(), ".xml") {
				if f, err := mp.NewFsFile("maps/"+p.Name(), filepath.Join(app.config.DataDir(), "maps", p.Name())); err == nil {
					app.logger.Debug("add " + p.Name())

					res = append(res, f)
				}
			}
		}
	}

	return res
}

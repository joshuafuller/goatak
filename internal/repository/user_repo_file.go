package repository

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	"github.com/kdudkov/goatak/pkg/model"
)

var _ AuthRepository = &AdminMemRepository{}

type AdminMemRepository struct {
	userFile string
	logger   *slog.Logger
	users    map[string]*model.Device

	watcher *fsnotify.Watcher

	mx sync.RWMutex
}

func NewAdminMemRepo() *AdminMemRepository {
	um := &AdminMemRepository{
		logger: slog.Default().With("logger", "AdmManager"),
		users:  make(map[string]*model.Device),
		mx:     sync.RWMutex{},
	}

	if err := um.loadUsersFile(); err != nil {
		um.logger.Error("error loading users file", slog.Any("error", err))
	}

	if len(um.users) == 0 {
		um.logger.Info("no valid users found - create one")

		user := &model.Device{
			Login: "user",
		}

		_ = user.SetPassword("test")
		um.users["user"] = user
	}

	return um
}

func (r *AdminMemRepository) loadUsersFile() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if _, err := os.Lstat(r.userFile); os.IsNotExist(err) {
		// create empty file
		f, err := os.Create(r.userFile)
		if err != nil {
			return err
		}

		return f.Close()
	}

	dat, err := os.ReadFile(r.userFile)
	if err != nil {
		return err
	}

	users := make([]*model.Device, 0)

	if err := yaml.Unmarshal(dat, &users); err != nil {
		return err
	}

	r.users = make(map[string]*model.Device)

	for _, user := range users {
		if user.Login != "" {
			r.users[user.Login] = user
		}
	}

	return nil
}

func (r *AdminMemRepository) Start() error {
	var err error
	r.watcher, err = fsnotify.NewWatcher()

	if err != nil {
		return err
	}

	if err := r.watcher.Add(r.userFile); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-r.watcher.Events:
				if !ok {
					return
				}

				r.logger.Debug(fmt.Sprintf("event: %v", event))

				if event.Has(fsnotify.Write) && event.Name == r.userFile {
					r.logger.Info("users file is modified, reloading")

					if err := r.loadUsersFile(); err != nil {
						r.logger.Error("error", slog.Any("error", err))
					}
				}
			case err, ok := <-r.watcher.Errors:
				if !ok {
					return
				}

				r.logger.Error("error", slog.Any("error", err))
			}
		}
	}()

	return nil
}

func (r *AdminMemRepository) Stop() {
	if r.watcher != nil {
		_ = r.watcher.Close()
	}
}

func (r *AdminMemRepository) CheckAuth(username, password string) bool {
	r.mx.RLock()
	defer r.mx.RUnlock()

	if user, ok := r.users[username]; ok {
		return user.IsGood() && user.CheckPassword(password)
	}

	return false
}

func (r *AdminMemRepository) IsValid(username, sn string) bool {
	r.mx.RLock()
	defer r.mx.RUnlock()
	if u, ok := r.users[username]; ok {
		return u.IsGood()
	}

	return false
}

func (r *AdminMemRepository) Get(username string) *model.Device {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.users[username]
}

func (r *AdminMemRepository) SaveSignInfo(username string, uid, sn string, till time.Time) {
	// no-op
}

func (r *AdminMemRepository) SaveConnectInfo(username string, uid, sn string) {
	// no-op
}

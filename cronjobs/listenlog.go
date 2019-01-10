package cronjobs

import (
	"context"

	"github.com/R-a-dio/valkyrie/database"
	"github.com/R-a-dio/valkyrie/engine"
	"github.com/R-a-dio/valkyrie/rpc/manager"
)

const insertLog = `INSERT INTO listenlog (listeners, dj) VALUES (?, ?);`

// ListenLog fetches the listener count from the manager and inserts a line into
// the listenlog table.
func ListenLog(e *engine.Engine) error {
	m := e.Conf().Manager.TwirpClient()

	status, err := m.Status(context.TODO(), &manager.StatusRequest{})
	if err != nil {
		return err
	}

	h := database.Handle(context.TODO(), e.DB)
	_, err = h.Exec(insertLog, status.ListenerInfo.Listeners, status.User.Id)
	if err != nil {
		return err
	}

	return nil
}
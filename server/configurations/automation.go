package configurations

import (
	"fmt"

	"github.com/pbloigu/gonfig/server/database"
	"github.com/rs/zerolog/log"
)

func (c *c) UpdateStatusChangeTrigger(appId string, t StatusChangeTrigger) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
            DELETE FROM Action
            WHERE
                status_change_trigger_id = (
                    SELECT id FROM
                    StatusTrigger
                    WHERE application_id = ?
                )`, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not delete old actions.")
			return nil, err
		}

		var sId int

		r, err := dba.Query(`
            SELECT id
            FROM StatusTrigger
            WHERE application_id = ?
        `, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not fetch trigger id.")
		}
		defer r.Close()
		if r.Next() {
			err = r.Scan(&sId)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not fetch trigger id.")
			}
		}

		for _, a := range t.Actions {
			_, err := dba.Exec(`
                INSERT INTO Action (description, script, status_change_trigger_id)
                VALUES (?, ?, ?)    
            `, a.Description, a.Script, sId)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not insert action.")
				return nil, err
			}
		}
		c.statusTriggers.Swap(appId, t)
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not update trigger..")
	}
}

func (c *c) UpdateCronTrigger(cr CronTrigger) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
            DELETE FROM Action
            WHERE
                cron_trigger_id = ?`, cr.Id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not delete old actions.")
			return nil, err
		}

		for _, a := range cr.Actions {
			_, err := dba.Exec(`
                INSERT INTO Action (description, script, cron_trigger_id)
                VALUES (?, ?, ?)    
            `, a.Description, a.Script, cr.Id)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not insert action.")
				return nil, err
			}
		}
		_, err = dba.Exec(`
			UPDATE CronTrigger
			SET description = ?
				expression = ?
			WHERE id = ?
		`, cr.Description, cr.CronExpression, cr.Id)

		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not update cron trigger.")
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not list update trigger..")
	}
}

func (c *c) GetStatusChangeActions(appId string) []Action {
	if res, ok := c.statusTriggers.Load(appId); ok {
		return res.(StatusChangeTrigger).Actions
	} else {
		return []Action{}
	}
}

func (c *c) GetStatusChangeTrigger(appId string) StatusChangeTrigger {
	tr, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		tr := StatusChangeTrigger{}
		r, err := dba.Query(`
            SELECT
                id
            FROM StatusTrigger
            WHERE application_id = ?`, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not list triggers.")
			return nil, err
		}
		defer r.Close()
		if r.Next() {
			err = r.Scan(&tr.Id)
			if err != nil {
				log.Panic().AnErr("error", err).Msg("Database operation failed.")
			}
			a, err := c.listActions(dba, tr)
			if err != nil {
				return nil, err
			}
			tr.Actions = a
		}

		return tr, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not list series triggers.")
	}

	return tr.(StatusChangeTrigger)
}

func (c *c) GetCronTrigger(id int) CronTrigger {
	cr, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		cr := CronTrigger{}
		r, err := dba.Query(`
            SELECT
                id
            FROM CronTrigger
            WHERE id = ?`, id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not list triggers.")
			return nil, err
		}
		defer r.Close()
		if r.Next() {
			err = r.Scan(&cr.Id)
			if err != nil {
				log.Panic().AnErr("error", err).Msg("Database operation failed.")
			}
			a, err := c.listActions(dba, cr)
			if err != nil {
				return nil, err
			}
			cr.Actions = a
		}

		return cr, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not list series triggers.")
	}

	return cr.(CronTrigger)
}

func (c *c) ListCronTriggers() []CronTrigger {
	res, err := c.db.Context().Query(`
		SELECT 
			ct.id,
			ct.description,
			ct.expression,
			a.description,
			a.script
		FROM CronTrigger ct
		INNER JOIN Action a on a.cron_trigger_id = ct.id`)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not list cron triggers.")
	}
	defer res.Close()
	cts := make(map[int]CronTrigger, 0)
	type row struct {
		id     int
		descr  string
		expr   string
		adescr string
		script string
	}
	for res.Next() {
		r := row{}
		res.Scan(&r.id, &r.descr, &r.expr, &r.adescr, &r.script)
		if ct, ok := cts[r.id]; ok {
			ct.Actions = append(ct.Actions, Action{
				Description: r.adescr,
				Script:      r.script,
			})
		} else {
			ct = CronTrigger{
				Id:          r.id,
				Description: r.descr,
				Actions:     make([]Action, 0),
			}
			ct.Actions = append(ct.Actions, Action{
				Description: r.adescr,
				Script:      r.script,
			})
			cts[r.id] = ct
		}
	}

	result := make([]CronTrigger, 0)
	for _, ct := range cts {
		result = append(result, ct)
	}
	return result

}

func (c *c) ListSeriesTriggers(appId string) []SeriesTrigger {
	ms, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		ms := make([]SeriesTrigger, 0)
		r, err := dba.Query(`
            SELECT
                id,
                series_name
            FROM SeriesTrigger
            WHERE application_id = ?
            `, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not list triggers.")
			return nil, err
		}
		defer r.Close()
		for r.Next() {
			m := SeriesTrigger{}
			err = r.Scan(&m.Id, &m.SeriesName)
			if err != nil {
				log.Panic().AnErr("error", err).Msg("Database operation failed.")
			}
			a, err := c.listActions(dba, m)
			if err != nil {
				return nil, err
			}
			m.Actions = a
			ms = append(ms, m)
		}

		return ms, nil
	})

	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not list series triggers.")
	}

	return ms.([]SeriesTrigger)
}

func (c *c) PersistCronTrigger(cr CronTrigger) CronTrigger {
	crId, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
            INSERT INTO CronTrigger (description, expression)
            VALUES (?, ?)
        `, cr.Description, cr.CronExpression)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not insert cron trigger.")
			return nil, err
		}

		r, err := dba.Query(`
            SELECT id
            FROM CronTrigger
            WHERE expression = ?
        `, cr.CronExpression)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Cron trigger was not inserted.")
			return nil, err
		}
		defer r.Close()
		if !r.Next() {
			log.Error().AnErr("error", err).Msg("Cron trigger was not inserted.")
			return nil, err
		}
		var crId int
		if err = r.Scan(&crId); err != nil {
			log.Error().AnErr("error", err).Msg("Cron trigger was not inserted.")
			return nil, err
		}

		for _, a := range cr.Actions {
			_, err := dba.Exec(`
                INSERT INTO Action (description, script, cron_trigger_id)
                VALUES (?, ?, ?)
            `, a.Description, a.Script, crId)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not insert action.")
				return nil, err
			}
		}
		return crId, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not persist status change trigger.")
	}
	return c.GetCronTrigger(crId.(int))

}

func (c *c) PersistStatusChangeTrigger(appId string, st StatusChangeTrigger) StatusChangeTrigger {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
            INSERT INTO StatusTrigger (application_id)
            VALUES (?)
        `, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not insert status change trigger.")
			return nil, err
		}

		r, err := dba.Query(`
            SELECT id
            FROM StatusTrigger
            WHERE application_id = ?
        `, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Status change trigger was not inserted.")
			return nil, err
		}
		defer r.Close()
		if !r.Next() {
			log.Error().AnErr("error", err).Msg("Status change trigger was not inserted.")
			return nil, err
		}
		var stId int
		if err = r.Scan(&stId); err != nil {
			log.Error().AnErr("error", err).Msg("Status change trigger was not inserted.")
			return nil, err
		}

		for _, a := range st.Actions {
			_, err := dba.Exec(`
                INSERT INTO Action (description, script, status_change_trigger_id)
                VALUES (?, ?, ?)
            `, a.Description, a.Script, stId)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not insert action.")
				return nil, err
			}
		}
		c.statusTriggers.Store(appId, st)
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not persist status change trigger.")
	}

	return c.GetStatusChangeTrigger(appId)
}

func (c *c) PersistSeriesTrigger(appId string, st SeriesTrigger) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
            INSERT INTO SeriesTrigger (application_id, series_name)
            VALUES (?, ?)
        `, appId, st.SeriesName)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not insert series trigger.")
			return nil, err
		}
		r, err := dba.Query(`
            SELECT id
            FROM SeriesTrigger
            WHERE application_id = ?
            AND series_name = ?
        `, appId, st.SeriesName)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Series trigger was not inserted.")
			return nil, err
		}
		defer r.Close()
		if !r.Next() {
			log.Error().AnErr("error", err).Msg("Series trigger was not inserted.")
			return nil, err
		}
		var stId int
		if err = r.Scan(&stId); err != nil {
			log.Error().AnErr("error", err).Msg("Series trigger was not inserted.")
			return nil, err
		}
		for _, a := range st.Actions {
			_, err := dba.Exec(`
                INSERT INTO Action (description, script, series_trigger_id)
                VALUES (?, ?, ?)
            `, a.Description, a.Script, stId)
			if err != nil {
				log.Error().AnErr("error", err).Msg("Could not insert action.")
				return nil, err
			}
		}
		c.seriesTriggers.Store(fmt.Sprintf("%s:%s", appId, st.SeriesName), st)
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Could not persist series trigger.")
	}

}

func (c *c) DeleteStatusChangeTrigger(appId string) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
            DELETE FROM Action
            WHERE status_change_trigger_id = (
                SELECT id
                FROM StatusTrigger
                WHERE application_id = ?
            )
        `, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Deleting actions failed.")
			return nil, err
		}
		_, err = dba.Exec(`
            DELETE FROM StatusTrigger
            WHERE application_id = ?
        `, appId)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Deleting status change trigger failed.")
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Status change trigger deletion failed.")
	}
}

func (c *c) DeleteCronTrigger(id int) {
	_, err := c.db.DoInTransaction(func(dba database.Context) (any, error) {
		_, err := dba.Exec(`
            DELETE FROM Action
            WHERE cron_trigger_id = ?
        `, id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Deleting actions failed.")
			return nil, err
		}
		_, err = dba.Exec(`
            DELETE FROM CronTrigger
            WHERE id = ?
        `, id)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Deleting cron trigger failed.")
			return nil, err
		}
		// c.cronActions.remove()
		return nil, nil
	})
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Cron trigger deletion failed.")
	}
}

func (c *c) populateTriggerCaches() {

	r, err := c.db.Context().Query(`
		SELECT id
		FROM Application
	`)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Unable to list application.")
		return
	}
	defer r.Close()
	for r.Next() {
		var appId string
		err = r.Scan(&appId)
		if err != nil {
			log.Panic().AnErr("error", err).Msg("Unable to select application id.")
			return
		}
		log.Debug().Any("appId", appId).Msg("Found app.")
		c.apps.Store(appId, appId)
	}

	waiter := make(chan bool, 3)
	go func() {
		c.apps.Range(func(key, value any) bool {
			appId := key.(string)
			tr := c.GetStatusChangeTrigger(appId)
			c.statusTriggers.Store(appId, tr)
			return true
		})

		waiter <- true
	}()

	go func() {
		c.apps.Range(func(key, value any) bool {
			appId := key.(string)
			mts := c.ListSeriesTriggers(appId)
			for _, mt := range mts {
				c.seriesTriggers.Store(fmt.Sprintf("%s:%s", appId, mt.SeriesName), mt)
			}
			return true
		})

		waiter <- true
	}()

	for range 2 {
		<-waiter
	}
}

func (c *c) listActions(dba database.Context, anyTrigger any) ([]Action, error) {
	acts := make([]Action, 0)
	var column string
	var id int
	switch t := anyTrigger.(type) {
	case SeriesTrigger:
		{
			column = "series_trigger_id"
			id = t.Id
		}
	case CronTrigger:
		{
			column = "cron_trigger_id"
			id = t.Id
		}
	case StatusChangeTrigger:
		{
			column = "status_change_trigger_id"
			id = t.Id
		}
	default:
		{
			err := fmt.Errorf("unable to handle action of type %s", t)
			log.Error().AnErr("error", err).Msg("Could not list actions.")
			return nil, err
		}
	}

	r, err := dba.Query(fmt.Sprintf(`
		SELECT
			id,
			description,
			script
		FROM Action
		WHERE %s = ?
		ORDER BY description ASC
	`, column), id)

	if err != nil {
		log.Error().AnErr("error", err).Msg("Could not list actions.")
		return nil, err
	}
	defer r.Close()
	for r.Next() {
		a := Action{}
		err = r.Scan(&a.Id, &a.Description, &a.Script)
		if err != nil {
			log.Error().AnErr("error", err).Msg("Could not list actions.")
			return nil, err
		}
		acts = append(acts, a)
	}
	return acts, nil
}

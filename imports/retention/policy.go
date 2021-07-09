// retention is a self implemented data retention solution
// It's very simple, but to resolve our basic requirement.
//
// Mattermost job system memo v5.35
// Jobserver
//    Watch:     Poll and notify Works every 15 secs
//               Check the job DB and find any pending jobs
//               send the job to specific job channel( return from Worker.JobChannel())
//    Scheduler: Schedule the next exection time
//               Put the job in DB as pending status, this must be implemented in ScheduleJob() using Jobserver.CreateJob
//    Worker:    Execute job. Wait until any job put in JobChannel()
// 
// Initialization flow:
// Server cmd - run-server:
//      a.NewServer
//            s.initEntprise
//            s.initJobs
//      fakeapp.initServer
//            a.initEnterprise
//            a.initJobs
//                a.srv.jobs.initWorks()
//                    a.srv.jobs.MakeWatcher()
//                a.src.jobs.initSchedules()
//                s.runjobs
//                    s.js.StartWorkers()
//                         workers.start() + watch.start() 
//                    s.js.StartSchedulers()
package retention

import (
	"github.com/mattermost/mattermost-server/v5/app"
	"github.com/mattermost/mattermost-server/v5/einterfaces"
	// ejobs "github.com/mattermost/mattermost-server/v5/einterfaces/jobs"
	"github.com/mattermost/mattermost-server/v5/model"
)

type SimpleRetention struct {
	srv *app.Server
}

func init() {
	app.RegisterDataRetentionInterface(NewSimpleRetention)
}

func NewSimpleRetention(s *app.Server) einterfaces.DataRetentionInterface {
	return &SimpleRetention{
		srv: s,
	}
}


func (zs *SimpleRetention) GetPolicy() (*model.DataRetentionPolicy, *model.AppError) {
	return nil, nil
}


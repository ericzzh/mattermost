// retention is a self implemented data retention solution
// It's very simple, but hope to resolve our basic requirement.
// 
// *. Default rentention rule follows the config setting
// *. Specific rentention rule is defined in a server csv file
//       type:    team/channel/direct
//       id:      team/channel ID/usrid(direct message, mainly a bot)
//       name:    for human reading
//       period:  retention period
// *. Pinned and user saved messsages won't be deleted.
// *. Message and its thread is a whole, which means unless the last date of the last thread of a message is expired,
//    the whole chats won't be deleted.
// *. If all the file in a directory are deleted, the fold will be delete too

// #  Some idea: 
// #  Set a trash bin and then permanently deleted? 
//
// Mattermost job system memo v5.35
//    Jobserver: Like a platform providing tools
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


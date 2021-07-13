// retention is a self implemented data retention solution
// It's very simple, but hope to resolve our basic requirement.
//
// *. Default rentention rule follows the config setting
// *. Specific rentention rule is defined in a server csv file
//       type:    team/channel/direct
//       id:      team/channel ID/usrid(direct message, mainly a bot)
//       name:    for human reading
//       period:  retention period
// *. Pinned message won't be deleted.
// *. Message and its thread is a whole, which means unless the last date of the last thread of a message is expired,
//    the whole chats won't be deleted.
// *. If all the file in a directory are deleted, the fold will be delete too ** to do in the future, unless the performace is bad

// #  Some idea:
// #  Set a trash bin and then permanently deleted? -- can use the build-in delete(just mark)
//
// Implement notes:
// Use api to get information, DON't use database directly
// But cleanng work will do directly with database anyway.
// Investigate API, to find the database operation
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

const SIMPLE_RETENTION_KIND_TEAM = "Team"
const SIMPLE_RETENTION_KIND_CHANNEL = "Channel"
const SIMPLE_RETENTION_KIND_USER = "User"

type SimplePolicy struct {
	period   int64
	specific []SimpleSpecificPolicy
}

type SimpleSpecificPolicy struct {
	kind   string
	id     string
	name   string
	period int64 // 0: indefinetly, seconds, input is days, should convert to second
}

var policy SimplePolicy

func SetPolicy(p SimplePolicy) {
	policy = p
}

func GetPolicy() SimplePolicy {
	return policy
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

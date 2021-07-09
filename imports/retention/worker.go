package retention

import (
	// "net/http"

	"github.com/mattermost/mattermost-server/v5/app"
	"github.com/mattermost/mattermost-server/v5/jobs"
	// tjobs "github.com/mattermost/mattermost-server/v5/jobs/interfaces"
	ejobs "github.com/mattermost/mattermost-server/v5/einterfaces/jobs"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/shared/mlog"
)

const (
	JobName = "SimpleRetention"
)

type Worker struct {
	name      string
	stop      chan bool
	stopped   chan bool
	jobs      chan model.Job
	jobServer *jobs.JobServer
	srv       *app.Server
}

func init() {
	app.RegisterJobsDataRetentionJobInterface(func(s *app.Server) ejobs.DataRetentionJobInterface {
		return &SimpleRetentionImpl{s}
	})
}

type SimpleRetentionImpl struct {
	Srv *app.Server
}

func (m *SimpleRetentionImpl) MakeWorker() model.Worker {
	worker := Worker{
		name:      JobName,
		stop:      make(chan bool, 1),
		stopped:   make(chan bool, 1),
		jobs:      make(chan model.Job),
		jobServer: m.Srv.Jobs,
		srv:       m.Srv,
	}
	return &worker
}

func (worker *Worker) Run() {
	mlog.Debug("***************ZZH************** Worker started", mlog.String("worker", worker.name))

	defer func() {
		mlog.Debug("***************ZZH************** Worker finished", mlog.String("worker", worker.name))
		worker.stopped <- true
	}()

	for {
		select {
		case <-worker.stop:
			mlog.Debug("***************ZZH************** Worker received stop signal", mlog.String("worker", worker.name))
			return
		case job := <-worker.jobs:
			mlog.Debug("***************ZZH************** Worker received a new candidate job.", mlog.String("worker", worker.name))
			worker.DoJob(&job)
		}
	}
}

func (worker *Worker) Stop() {
	mlog.Debug("***************ZZH************** Worker stopping", mlog.String("worker", worker.name))
	worker.stop <- true
	<-worker.stopped
}

func (worker *Worker) JobChannel() chan<- model.Job {
	return worker.jobs
}

func (worker *Worker) DoJob(job *model.Job) {
	if claimed, err := worker.jobServer.ClaimJob(job); err != nil {
		mlog.Warn("***************ZZH************** Worker experienced an error while trying to claim job",
			mlog.String("***************ZZH************** Worker", worker.name),
			mlog.String("job_id", job.Id),
			mlog.String("error", err.Error()))
		return
	} else if !claimed {
		return
	}

	// count, err := worker.srv.Srv().Store.User().Count(model.UserCountOptions{IncludeDeleted: false})

	// if err != nil {
	// 	mlog.Error("***************ZZH************** Worker: Failed to get active user count", mlog.String("worker", worker.name), mlog.String("job_id", job.Id), mlog.String("error", err.Error()))
	// 	worker.setJobError(job, model.NewAppError("DoJob", "app.user.get_total_users_count.app_error", nil, err.Error(), http.StatusInternalServerError))
	// 	return
	// }

	// if worker.srv.Metrics() != nil {
	// 	worker.srv.Metrics().ObserveEnabledUsers(count)
	// }

	mlog.Info("***************ZZH************** Worker: Job is complete", mlog.String("worker", worker.name), mlog.String("job_id", job.Id))
	worker.setJobSuccess(job)
}

func (worker *Worker) setJobSuccess(job *model.Job) {
	if err := worker.srv.Jobs.SetJobSuccess(job); err != nil {
		mlog.Error("***************ZZH************** Worker: Failed to set success for job", mlog.String("worker", worker.name), mlog.String("job_id", job.Id), mlog.String("error", err.Error()))
		worker.setJobError(job, err)
	}
}

func (worker *Worker) setJobError(job *model.Job, appError *model.AppError) {
	if err := worker.srv.Jobs.SetJobError(job, appError); err != nil {
		mlog.Error("***************ZZH************** Worker: Failed to set job error", mlog.String("worker", worker.name), mlog.String("job_id", job.Id), mlog.String("error", err.Error()))
	}
}

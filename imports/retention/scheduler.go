package retention

import (
	"time"

	"github.com/mattermost/mattermost-server/v5/app"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	SchedFreqMinutes = 5
)

type Scheduler struct {
	Srv *app.Server
}

func (m *SimpleRetentionImpl) MakeScheduler() model.Scheduler {
	return &Scheduler{m.Srv}
}

func (scheduler *Scheduler) Name() string {
	return JobName + "Scheduler"
}

func (scheduler *Scheduler) JobType() string {
	return model.JOB_TYPE_DATA_RETENTION
}

func (scheduler *Scheduler) Enabled(cfg *model.Config) bool {
	return *cfg.DataRetentionSettings.EnableFileDeletion || *cfg.DataRetentionSettings.EnableMessageDeletion

}

func (scheduler *Scheduler) NextScheduleTime(cfg *model.Config, now time.Time, pendingJobs bool, lastSuccessfulJob *model.Job) *time.Time {
	nextTime := time.Now().Add(SchedFreqMinutes * time.Second)
	return &nextTime
}

func (scheduler *Scheduler) ScheduleJob(cfg *model.Config, pendingJobs bool, lastSuccessfulJob *model.Job) (*model.Job, *model.AppError) {
	data := map[string]string{}

	job, err := scheduler.Srv.Jobs.CreateJob(model.JOB_TYPE_DATA_RETENTION, data)
	if err != nil {
		return nil, err
	}
	return job, nil
}

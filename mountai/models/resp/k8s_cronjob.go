package resp

type DescribeCronJobResp struct {
	Status CronJobStatus `json:"status"`
	Events []Event       `json:"events"`
}

type CronJobStatus struct {
	Name             string `json:"name"`
	LastScheduleTime string `json:"last_schedule_time"`
}

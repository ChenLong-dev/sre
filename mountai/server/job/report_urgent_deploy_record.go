package job

import (
	"rulai/service"

	framework "gitlab.shanhai.int/sre/app-framework"
)

func ReportUrgentDeployRecordServer() framework.ServerInterface {
	svr := new(framework.JobServer)
	svr.SetJob("report_urgent_deploy_record", service.SVC.ReportUrgentDeployRecord)

	return svr
}

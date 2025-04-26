package req

type DeleteJob struct {
	Clusterame string `binding:"required" json:"clusterName"`
	EnvName    string `binding:"required" json:"envName"`
	Namespace  string `binding:"required" json:"namespace"`
	Podname    string `binding:"required" json:"podName"`
	AppId      string `binding:"required" json:"appId"`
}

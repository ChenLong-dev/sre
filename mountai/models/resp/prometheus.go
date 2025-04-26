package resp

const (
	PrometheusStatusSuccess = "success"
	PrometheusStatusError   = "error"
)

type QueryPrometheusResp struct {
	Status    string                   `json:"status"`
	ErrorType string                   `json:"errorType"`
	Error     string                   `json:"error"`
	Data      *QueryPrometheusDataResp `json:"data"`
}

type QueryPrometheusDataResp struct {
	ResultType string `json:"resultType"`
	Result     []struct {
		Metric map[string]string `json:"metric"`
		Value  []interface{}     `json:"value"`
	} `json:"result"`
}

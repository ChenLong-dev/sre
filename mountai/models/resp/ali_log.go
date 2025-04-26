package resp

const (
	// TTLPERMANNENT 表示永久保存
	// https://help.aliyun.com/document_detail/48990.html?spm=5176.10695662.1996646101.searchclickresult.6f86278dY9tww2
	TTLPERMANNENT int = 3650
)

// AliGetLogStoreDetailResp 日志仓库详细信息
type AliGetLogStoreDetailResp struct {
	LogStoreName string `json:"log_store_name"`
	// 数据保存时间
	TTL int `json:"ttl"`
}

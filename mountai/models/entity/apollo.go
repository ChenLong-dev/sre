package entity

type ApolloNamespace struct {
	ID            int    `json:"id" gorm:"column:Id"`
	AppID         string `json:"app_id" gorm:"column:AppId"`
	ClusterName   string `json:"cluster_name" gorm:"column:ClusterName"`
	NamespaceName string `json:"namespace_name" gorm:"column:NamespaceName"`
}

func (*ApolloNamespace) TableName() string {
	return "namespace"
}

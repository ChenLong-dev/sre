package dao

func (d *Dao) GetApolloConf(namespace string) map[string]interface{} {
	// 获取namespace的所有配置
	allConf := d.ApolloClient.Get(namespace)

	return allConf
}

package service

// 获取apollo配置
func (s *Service) GetApolloConf(namespace string) map[string]interface{} {
	allConf := s.dao.GetApolloConf(namespace)

	return allConf
}

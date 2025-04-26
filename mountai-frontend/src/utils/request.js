/**
 * request 网络请求工具
 * 更详细的 api 文档: https://github.com/umijs/umi-request
 */
import { extend } from 'umi-request';
import { notification, message } from 'antd';
import { getPageQuery } from '@/utils/utils';
/**
 * 异常处理程序
 */

const errorHandler = (error) => {
  const { response, data } = error;

  if (response && response.status) {
    const { status } = response;
    if (status && status === 401) {
      const { redirect } = getPageQuery(); // Note: There may be security issues, please note
      localStorage.removeItem('ams-user-authority');
      if (window.location.pathname !== '/user/login' && !redirect) {
        message.error('登录已过期');
        window.location.href = `/user/login?redirect=${window.location.href}`;
      }
    }

    return data;
  }

  if (!response) {
    notification.error({
      description: '您的网络发生异常，无法连接服务器',
      message: '网络异常',
    });
  }

  throw error;
};
/**
 * 配置request请求时的默认参数
 */

const request = extend({
  errorHandler,
  // 默认错误处理
  // 每个接口必传token
  // headers: { 'Authorization': localStorage.getItem('ams-user-authority')  ? `Bearer ${JSON.parse(localStorage.getItem('ams-user-authority')).token}` :null},
  // credentials: 'include', // 默认请求是否带上cookie
});
request.interceptors.request.use((url, options) => {
  return {
    url,
    options: {
      ...options,
      headers: {
        ...options.headers,
        Authorization: localStorage.getItem('ams-user-authority')
          ? `Bearer ${
              JSON.parse(localStorage.getItem('ams-user-authority')).token
            }`
          : null,
      },
    },
  };
});
export default request;

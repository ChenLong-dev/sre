import { stringify } from 'querystring';
import { history } from 'umi';
import { fakeAccountLogin } from '@/services/login';
import { getPageQuery } from '@/utils/utils';
import { reloadAuthorized } from '@/utils/Authorized';
import { message } from 'antd';

const Model = {
  namespace: 'login',
  state: {
    status: undefined,
  },
  effects: {
    *login({ payload }, { call, put }) {
      try {
        const response = yield call(fakeAccountLogin, payload);

        yield put({
          type: 'changeLoginStatus',
          payload: response,
        });

        // Login successfully
        // 登录后返回
        if (response && response.token) {
          localStorage.clear();
          localStorage.setItem('ams-user-authority', JSON.stringify(response));

          reloadAuthorized();
          const urlParams = new URL(window.location.href);
          const params = getPageQuery();
          let { redirect } = params;

          if (redirect) {
            const redirectUrlParams = new URL(redirect);

            if (redirectUrlParams.origin === urlParams.origin) {
              redirect = redirect.substr(urlParams.origin.length);

              if (redirect.match(/^\/.*#/)) {
                redirect = redirect.substr(redirect.indexOf('#') + 1);
              }
            } else {
              redirect = '/';
              return;
            }
          }

          setTimeout(() => {
            window.location.href = redirect || '/';
          }, 0);
        } else {
          message.error('登录失败');
        }
      } catch (error) {
        message.error(error.data?.errmsg || error.message);
      }
    },

    logout() {
      const { redirect } = getPageQuery(); // Note: There may be security issues, please note
      localStorage.removeItem('ams-user-authority');
      if (window.location.pathname !== '/user/login' && !redirect) {
        history.replace({
          pathname: '/user/login',
          search: stringify({
            redirect: window.location.href,
          }),
        });
      }
    },
  },
  reducers: {
    changeLoginStatus(state, { payload }) {
      return { ...state, status: payload.status, type: payload.type };
    },
  },
};
export default Model;

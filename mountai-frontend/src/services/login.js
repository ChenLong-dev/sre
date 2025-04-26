import ServiceManager from './ServiceManager';

export function fakeAccountLogin(params) {
  return ServiceManager.post('/amsapi/api/v1/login', {
    data: { username: params.userName, password: params.password },
  });
}

export function getFakeCaptcha(mobile) {
  return ServiceManager.get(`/api/login/captcha?mobile=${mobile}`);
}

import ServiceManager from './ServiceManager';

export function query() {
  return ServiceManager.get('/api/users');
}
// 拿登录信息
export async function queryCurrent() {
  return JSON.parse(localStorage.getItem('ams-user-authority'));
}
export function queryNotices() {
  return ServiceManager.get('/api/notices');
}

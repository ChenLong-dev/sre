import ServiceManager from './ServiceManager';

export default function getUsers(keyword) {
  return ServiceManager.get('/amsapi/api/v1/users', {
    params: {
      k: keyword,
      limit: 50,
      page: 1,
    },
  });
}

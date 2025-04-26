import ServiceManager from './ServiceManager';

export function queryAllTeams() {
  const data = { limit: 50, page: 1 };
  return ServiceManager.get('/amsapi/api/v1/teams', {
    params: data,
  });
}

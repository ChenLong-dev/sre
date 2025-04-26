import querystring from 'querystring';

import ServiceManager from '../../services/ServiceManager';

export async function queryProjectNotice() {
  return ServiceManager.get('/api/project/notice');
}

export async function queryActivities() {
  return ServiceManager.get('/api/activities');
}
export async function fakeChartData() {
  return ServiceManager.get('/api/fake_chart_data');
}
export async function queryCurrent() {
  return ServiceManager.get('/api/currentUser');
}

// teamlist
export async function queryTeam() {
  // console.log('ganjiuwanle');
  const data = { limit: 50, page: 1 };

  return ServiceManager.get('/amsapi/api/v1/teams', {
    params: data,
  });
}

// 查询最新使用app
export async function queryUsedProjects() {
  let userId = null;
  if (
    !localStorage.getItem('ams-user-authority') &&
    !JSON.parse(localStorage.getItem('ams-user-authority')).id
  ) {
    return [];
  }
  userId = JSON.parse(localStorage.getItem('ams-user-authority')).id;

  return ServiceManager.get(`/amsapi/api/v1/users/${userId}/used_projects`);
}

export async function queryActivitiesLog() {
  const userInfo = JSON.parse(localStorage.getItem('ams-user-authority')) || {};
  const userId = userInfo?.id || '';
  return ServiceManager.get(
    `/amsapi/api/v1/activities?page=1&limit=10&operator_id=${userId}`
  );
}

// 获取收藏列表
export async function getCollectData(query) {
  return ServiceManager.get(
    `/amsapi/api/v1/fav_projects?${querystring.stringify(query)}`
  );
}

// 取消收藏项目
export const uncollectProject = async (project_id) => {
  return ServiceManager.delete(`/amsapi/api/v1/fav_projects/${project_id}`);
};

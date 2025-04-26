import ServiceManager from '../../../services/ServiceManager';

export function postProject(params) {
  return ServiceManager.post('/amsapi/api/v1/projects', {
    data: params,
  });
}

export async function updateProject(params) {
  // params.labels=  ['internal'];
  // console.log('更新params==>', params);

  return ServiceManager.put(`/amsapi/api/v1/projects/${params.id}`, {
    data: params,
  });
}

export async function getProjectInfo(projectId) {
  return ServiceManager.get(`/amsapi/api/v1/projects/${projectId}`);
}

export async function getGitInfobyid(git_id) {
  return ServiceManager.get(`/amsapi/api/v1/git/projects/${git_id}`);
}

export async function getTeams() {
  const data = { limit: 50, page: 1 };

  return ServiceManager.get('/amsapi/api/v1/teams', {
    params: data,
  });
}

// 获取所有对应项目label映射关系
export const getAllProjectLabels = async () => {
  return ServiceManager.get('/amsapi/api/v1/project_labels');
};

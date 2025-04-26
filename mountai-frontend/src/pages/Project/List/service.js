import { pick } from 'lodash';
import ServiceManager from '../../../services/ServiceManager';

export async function queryProject(params) {
  try {
    const result = await ServiceManager.get(`/amsapi/api/v1/projects`, {
      params: {
        ...pick(params, [
          'team_id',
          'owner_id',
          'language',
          'keyword',
          'labels',
        ]),
        limit: params.pageSize,
        page: params.current,
      },
    });
    return {
      current: result.page,
      data: result.list,
      pageSize: result.limit,
      success: true,
      total: result.count,
    };
  } catch {
    return {
      current: 1,
      data: [],
      pageSize: 0,
      success: true,
      total: 0,
    };
  }
}

export async function removeProject(params) {
  return ServiceManager.delete(`/amsapi/api/v1/projects/${params.id}`);
}
export async function addProject(params) {
  return ServiceManager.post('/amsapi/api/v1/projects', {
    data: { ...params, method: 'post' },
  });
}

export async function updateProject(params) {
  return ServiceManager.post('/api/projects', {
    data: { ...params, method: 'update' },
  });
}
// 获取所有对应项目label映射关系
export const getAllProjectLabels = async () => {
  return ServiceManager.get('/amsapi/api/v1/project_labels');
};

// 收藏项目
export const collectProject = async (project_id) => {
  return ServiceManager.post(`/amsapi/api/v1/fav_projects`, {
    data: { project_id },
  });
};

// 取消收藏项目
export const uncollectProject = async (project_id) => {
  return ServiceManager.delete(`/amsapi/api/v1/fav_projects/${project_id}`);
};

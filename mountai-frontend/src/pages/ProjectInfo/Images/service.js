import ServiceManager from '../../../services/ServiceManager';

export async function queryRule(params) {
  try {
    const result = await ServiceManager.get(
      `/amsapi/api/v1/projects/${params.projectId}/images/jobs`,
      {
        params: { limit: params.pageSize, page: params.current },
      }
    );
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

export async function removeRule(params) {
  return ServiceManager.delete(`/amsapi/api/v1/teams/${params.id}`);
}
export async function addRule(params) {
  return ServiceManager.post('/amsapi/api/v1/teams', {
    data: { ...params, method: 'post' },
  });
}

export async function updateRule(params) {
  return ServiceManager.post('/amsapi/api/v1/teams', {
    data: { ...params, method: 'update' },
  });
}

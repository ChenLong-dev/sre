import ServiceManager from '../../../services/ServiceManager';

export async function queryActivities(params) {
  try {
    const result = await ServiceManager.get(`/amsapi/api/v1/activities`, {
      params: {
        limit: params.pageSize,
        page: params.current,
        env_name: params.env_name || null,
        action: params.action || null,
        app_name: params.app_name || null,
        app_type: params.app_type || null,
        project_id: params.projectId || null,
      },
    });
    return {
      current: result.page || 1,
      data: result.list,
      pageSize: result.limit || 20,
      success: true,
      total: result.count || 500,
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

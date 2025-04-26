import { pick } from 'lodash';
import ServiceManager from '../../../services/ServiceManager';

export async function queryActivities(params) {
  try {
    const result = await ServiceManager.get(`/amsapi/api/v1/activities`, {
      params: {
        ...pick(params, ['env_name', 'app_name', 'app_type', 'is_fav']),
        project_id: params.projectId || null,
        limit: params.pageSize,
        page: params.current,
      },
    });
    return {
      current: result.page || 1,
      data: result.list,
      pageSize: result.limit || params.pageSize,
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

export function removeRule(params) {
  return ServiceManager.delete(`/amsapi/api/v1/teams/${params.id}`);
}

// import ServiceManager.get from 'umi-ServiceManager.get';
import querystring from 'querystring';
import { isNil, omitBy, pick } from 'lodash';
import ServiceManager from '../../../services/ServiceManager';

export async function fakeSubmitForm(params) {
  return ServiceManager.post('/amsapi/api/v1/projects', {
    data: { ...params, labels: ['internal'] },
  });
}

/**
 *
 * 创建镜像
 * @param {*} params
 */
export async function buildImage(params) {
  return ServiceManager.post(
    `/amsapi/api/v1/projects/${params.projectId}/images/jobs`,
    {
      data: pick(params, [
        'commit_id',
        'build_arg',
        'branch_name',
        'description',
        'build_args_template_id',
      ]),
    }
  );
}

/**
 *  删除构建任务
 * @param {*} projectId
 */
export async function deleteImagesBuild(projectId, buildId) {
  return ServiceManager.delete(
    `/amsapi/api/v1/projects/${projectId}/images/jobs/${buildId}`
  );
}

/**
 *
 * 提交创建应用
 * @param {*} params
 */
export async function createApp(params) {
  return ServiceManager.post(`/amsapi/api/v1/apps`, {
    data: {
      ...pick(params, [
        'name',
        'type',
        'project_id',
        'sentry_project_slug',
        'description',
        'service_type',
        'enable_istio',
      ]),
      enable_branch_change_notification:
        params.enable_branch_change_notification || false,
    },
  });
}

// 为应用添加sentry
export async function postSentry(appid) {
  return ServiceManager.post(`/amsapi/api/v1/apps/${appid}/sentry`);
}

// 应用批量操作
export async function postBatchAction(data) {
  return ServiceManager.post('/amsapi/api/v1/tasks/batch', {
    data,
  });
}

/**
 *
 * 删除app
 * @param {*} params
 */
export async function deleteApp(appid, data) {
  return ServiceManager.delete(`/amsapi/api/v1/apps/${appid}`, {
    data,
  });
}

/**
 *
 * 更新app
 * @param {*} params
 */
export async function updateApp(params) {
  return ServiceManager.put(`/amsapi/api/v1/apps/${params.id}`, {
    data: params,
  });
}

export async function getProjectInfo(projectId) {
  return ServiceManager.get(`/amsapi/api/v1/projects/${projectId}`);
}

export async function getProjectImagesInfo(projectId) {
  return ServiceManager.get(
    `/amsapi/api/v1/projects/${projectId}/images/jobs?limit=2&page=1`
  );
}

export async function getProjectSuccessImagesInfo(projectId, branchName) {
  return ServiceManager.get(
    `/amsapi/api/v1/projects/${projectId}/images/jobs?limit=2&page=1&branch_name=${branchName}&status=SUCCESS`
  );
}

export async function getProjectApplication(query) {
  return ServiceManager.get(
    `/amsapi/api/v1/apps?${querystring.stringify(query)}`
  );
}

export async function getProjectimageTag(projectId) {
  return ServiceManager.get(
    `/amsapi/api/v1/projects/${projectId}/images/tags?limit=50&page=1`
  );
}

/**
 *  获取配置  http://yapi.qingtingfm.com/project/547/interface/api/13340
 * @param {*} projectId
 */
export async function getProjectConfig(projectId, envname) {
  return ServiceManager.get(
    `/amsapi/api/v1/projects/${projectId}/config?env_name=${envname}&format_type=json`
  );
}

/**
 *  获取信息详情 api/v1/tasks/latest?env_name=prd&app_id=5f277192d92974b706524a09
 * @param {*} appId
 * @param {*} envname
 */
export async function getLatestTasks(
  appId,
  envname,
  clusterName,
  version = '',
  ignoreStatus = false
) {
  let params = {
    version,
    app_id: appId,
    env_name: envname,
    ignore_status: ignoreStatus,
    cluster_name: clusterName,
  };

  params = omitBy(params, isNil);

  return ServiceManager.get(
    `/amsapi/api/v1/tasks/latest?${querystring.stringify(params)}`
  );
}

// 获取创建发布任务高级选项下拉内容
export const getNodeLabels = async () => {
  return ServiceManager.get('/amsapi/api/v1/node_labels');
};

/**
 *  创建任务
 * @param {*} params
 */
export async function postTasks({
  appId,
  envName,
  action,
  version,
  param,
  clusterName,
  namespace,
  description,
  ignoreExpectedBranch,
}) {
  return ServiceManager.post(`/amsapi/api/v1/tasks`, {
    data: {
      action,
      version,
      param,
      namespace,
      description,
      app_id: appId,
      env_name: envName,
      cluster_name: clusterName,
      ignore_expected_branch: ignoreExpectedBranch,
    },
  });
}

/**
 *  获取信息详情
 * @param {*} appId
 * @param {*} envname
 */
export async function getApplicationInfo({ appId, envName, clusterName }) {
  return ServiceManager.get(
    `/amsapi/api/v1/apps/${appId}?env_name=${envName}&cluster_name=${clusterName}`
  );
}

export async function getProjectClusters(projectId) {
  return ServiceManager.get(`/amsapi/api/v1/projects/${projectId}/clusters`);
}

// 获取应用系统提示
export async function getApplicationTips(appId, envName, clusterName) {
  return ServiceManager.get(
    `/amsapi/api/v1/apps/${appId}/tips?env_name=${envName}&cluster_name=${clusterName}`
  );
};

// 创建Go项目pprof
export const createPodPprofData = (version, pod_name, query, body) => {
  const options = { data: body };
  if (body.action === 'download') options.responseType = 'blob';
  return ServiceManager.post(
    `/amsapi/api/v1/running_status/${version}/pods/${pod_name}/pprof?${querystring.stringify(
      query
    )}`,
    options
  );
};

/**
 *  暂停镜像部署
 */

export async function setImageStatusStop(task_id, suspend) {
  return ServiceManager.put(`/amsapi/api/v1/tasks/${task_id}`, {
    data: { suspend },
  });
}

/**
 *  创建任务
 * @param {*} params
 */
export async function postAppCorrectName(params) {
  const { appid, name } = params;
  return ServiceManager.post(`/amsapi/api/v1/apps/${appid}/correct_name`, {
    data: { name },
  });
}

export async function getRunningtask({ taskId }) {
  return ServiceManager.get(`/amsapi/api/v1/tasks/${taskId}`);
}

// 用于获取deployment_pods
export async function getRunningStatus({
  version,
  envName,
  appId,
  clusterName,
  namespace,
}) {
  return ServiceManager.get(`/amsapi/api/v1/running_status/${version}`, {
    params: {
      env_name: envName,
      app_id: appId,
      cluster_name: clusterName,
      namespace,
    },
  });
}

// 获取pod的describe信息
export async function getRunningPodLog({
  version,
  envName,
  podName,
  clusterName,
  namespace,
  containerName
}) {
  return ServiceManager.get(
    `/amsapi/api/v1/running_status/${version}/pods/${podName}/logs`,
    {
      params: {
        namespace,
        env_name: envName,
        cluster_name: clusterName,
        container_name: containerName
      },
    }
  );
}

// 获取应用的describe信息
export const getApplicationDescribe = async (version, query) => {
  return ServiceManager.get(
    `/amsapi/api/v1/running_status/${version}/description?${querystring.stringify(
      query
    )}`
  );
};

// 获取pod describe信息
export const getApplicationPodDescribe = async ({
  version,
  podName,
  envName,
  clusterName,
  namespace,
  containerName
}) => {
  return ServiceManager.get(
    `/amsapi/api/v1/running_status/${version}/pods/${podName}/description`,
    {
      params: {
        namespace,
        env_name: envName,
        cluster_name: clusterName,
        container_name: containerName
      },
    }
  );
};

// 获取项目是否创建过ci流程
export const getProjectCIStatus = async (projectId) => {
  return ServiceManager.get(`/amsapi/api/v1/projects/${projectId}/ci_job`);
};

// 为项目创建CI流程
export const setProjectCI = async (projectId, value) => {
  return ServiceManager.post(`/amsapi/api/v1/projects/${projectId}/ci_job`, {
    data: value,
  });
};

// 为项目修改CI流程
export const putProjectCIStatus = async (projectId, value) => {
  return ServiceManager.put(`/amsapi/api/v1/projects/${projectId}/ci_job`, {
    data: value,
  });
};

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

export const getProjectVariables = async (projectId) => {
  return ServiceManager.get(
    `/amsapi/api/v1/variables?project_id=${projectId}&limit=20&page=1&type=1`
  );
};

export const addProjectVariable = async (projectId, key, value) => {
  return ServiceManager.post('/amsapi/api/v1/variables', {
    data: {
      key,
      value,
      project_id: projectId,
      type: 1,
    },
  });
};

export const editProjectVariable = async (variableId, key, value) => {
  return ServiceManager.put(`/amsapi/api/v1/variables/${variableId}`, {
    data: {
      key,
      value,
    },
  });
};

export const deleteProjectVariable = async (variableId) => {
  return ServiceManager.delete(`/amsapi/api/v1/variables/${variableId}`);
};

/**
 *  订阅应用
 * @param {*} params
 */
export async function postUserSubscriptions(params) {
  return ServiceManager.post(`/amsapi/api/v1/user_subscriptions`, {
    data: params,
  });
}

/**
 *  取消订阅应用
 * @param {*} params
 */
export async function deleteUserSubscriptions(params) {
  return ServiceManager.delete(`/amsapi/api/v1/user_subscriptions`, {
    data: params,
  });
}

export function getMemberRole(projectId) {
  const userInfo = JSON.parse(localStorage.getItem('ams-user-authority')) || {};
  const userId = userInfo?.id || '';
  if (userId) {
    return ServiceManager.get(
      `/amsapi/api/v1/projects/${projectId}/members/${userId}/role`
    );
  }
  return null;
}

export function getAppClusters(id, query) {
  return ServiceManager.get(
    `/amsapi/api/v1/apps/${id}/clusters_with_workload?${querystring.stringify(
      query
    )}`
  );
}

export function getAppClusterWeights({ appId, env }) {
  return ServiceManager.get(
    `/amsapi/api/v1/apps/${appId}/cluster_weights?${querystring.stringify({
      env,
    })}`
  );
}

export function postAppClusterWeights({
  appId,
  env,
  clusterWeights,
  forceUpdateAll = false,
}) {
  return ServiceManager.post(`/amsapi/api/v1/apps/${appId}/cluster_weights`, {
    data: {
      env,
      cluster_weights: clusterWeights,
      force_update_all: forceUpdateAll,
    },
  });
}

/**
 *
 * 删除job
 * @param {*} params
 */
export async function deleteJob(data) {
  return ServiceManager.delete(`/amsapi/api/v1/jobs`, {
    data,
  });
}
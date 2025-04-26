import ServiceManager from './ServiceManager';

/**
 *
 *  http://yapi.qingtingfm.com/project/547/interface/api/13285
 * @param {*} projectId
 * @param {*} buildId
 */
export function getImageLog(projectId, buildId) {
  return ServiceManager.get(
    `/amsapi/api/v1/projects/${projectId}/images/jobs/${buildId}`
  );
}

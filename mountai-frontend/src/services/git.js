import ServiceManager from './ServiceManager';

export function queryGitInfo(projectId) {
  return ServiceManager.get(`/amsapi/api/v1/git/projects/${projectId}`);
}

export function queryGitBranch(projectId, keyword) {
  const urlParams = keyword
    ? `/amsapi/api/v1/git/projects/${projectId}/branch?keyword=${keyword}`
    : `/amsapi/api/v1/git/projects/${projectId}/branch`;
  return ServiceManager.get(urlParams);
}

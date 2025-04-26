import ServiceManager from '../../../services/ServiceManager';

export function getTemplates(teamId, page, limit) {
  return ServiceManager.get('/amsapi/api/v1/image_args_templates', {
    params: {
      page,
      limit,
      team_id: teamId,
    },
  });
}

export function addTemplate(teamId, name, content) {
  return ServiceManager.post('/amsapi/api/v1/image_args_templates', {
    data: {
      name,
      content,
      team_id: teamId,
    },
  });
}

export function editTemplate(templateId, name, content) {
  return ServiceManager.put(
    `/amsapi/api/v1/image_args_templates/${templateId}`,
    {
      data: {
        name,
        content,
      },
    }
  );
}

export function deleteTemplate(templateId) {
  return ServiceManager.delete(
    `/amsapi/api/v1/image_args_templates/${templateId}`
  );
}

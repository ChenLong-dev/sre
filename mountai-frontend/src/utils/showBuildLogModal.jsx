import React from 'react';
import { Modal } from 'antd';
import BuildLog from '@/components/BuildLog';

export default function showBuildLogModal(projectId, buildId) {
  Modal.info({
    title: '日志',
    width: '80vw',
    maskClosable: true,
    content: <BuildLog projectId={projectId} buildId={buildId} />,
  });
}

import React from 'react';
import { Modal, Collapse, message } from 'antd';
import { CopyOutlined } from '@ant-design/icons';
import copy from 'copy-to-clipboard';

// 最新配置-yaml文件复制功能
const YamlCopyList = ({
  filterInfo,
  yamlData,
  showYamlVisible,
  handleSetVisible,
}) => {
  const handleCopy = (item, data) => {
    copy(data);
    message.destroy();
    message.success(`已复制${item}文件`);
  };
  return (
    <Modal
      width="80%"
      title={
        <>
          最新配置
          <span style={{ color: 'red' }}>{`（${filterInfo.envname}）`}</span>
        </>
      }
      okText="确认"
      cancelText="取消"
      visible={showYamlVisible}
      onOk={() => handleSetVisible(false)}
      onCancel={() => handleSetVisible(false)}
    >
      <Collapse expandIconPosition="left">
        {Object.keys(yamlData || {}).map((item) => (
          <Collapse.Panel
            header={item}
            key={item}
            extra={
              <div
                style={{ color: '#1890ff' }}
                onClick={(event) => {
                  event.stopPropagation();
                  handleCopy(item, yamlData[item]);
                }}
              >
                <CopyOutlined />
                &nbsp;复制
              </div>
            }
          >
            <div
              style={{ whiteSpace: 'pre', cursor: 'pointer' }}
              onClick={() => handleCopy(item, yamlData[item])}
            >
              {yamlData[item]}
            </div>
          </Collapse.Panel>
        ))}
      </Collapse>
    </Modal>
  );
};

export default YamlCopyList;

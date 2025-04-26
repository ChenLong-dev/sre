import React, { useEffect, useState } from 'react';
import { message, Descriptions, Typography, Tag } from 'antd';
import {
  GitlabOutlined,
  ApiOutlined,
  FileTextOutlined,
  EditOutlined,
  CiCircleOutlined,
  StarFilled,
} from '@ant-design/icons';
import { queryGitInfo } from '@/services/git';
import { join, map } from 'lodash';
import styles from './style.less';
import { getProjectCIStatus } from '../service';
import EditCIProject from './EditCIProject';
import { hasDeveloperPermission } from '../utils/hasPermission';

const { Title } = Typography;

const ProjectInfo = ({
  data,
  allLabels,
  collect,
  memberRole,
  handleColllectProject,
  filterInfo,
}) => {
  const [isHighlightCiButton, setIsHighlightCiButton] = useState(false);
  const [isModalVisible, setIsModalVisible] = useState(false);

  const openGitLink = () => {
    queryGitInfo(data.id).then((res) => {
      if (res.web_url) {
        window.open(res.web_url);
      }
    });
  };

  const onOk = () => {
    setIsHighlightCiButton(true);
    setIsModalVisible(false);
  };

  useEffect(() => {
    (async () => {
      if (data) {
        try {
          const result = await getProjectCIStatus(data.id);
          setIsHighlightCiButton(!!result?.view_url);
        } catch (error) {
          if (error.code !== 1060006) {
            message.error(error.message);
          }
        }
      }
    })();
  }, [data]);

  return (
    <>
      {data && (
        <div className={styles.projectInfo}>
          <div>
            <Title level={3}>
              #{data.id} {data.name}
              {hasDeveloperPermission(memberRole, filterInfo.envname) && (
                <a
                  className={styles.editOutlined}
                  href={`/projects/update/${data.id}?redirect=/project/${data.id}/application`}
                >
                  <EditOutlined />
                </a>
              )}
              &nbsp;
              {/* <CiCircleOutlined
                style={
                  isHighlightCiButton ? { color: '#1890ff' } : { color: '#aaa' }
                }
                onClick={() => setIsModalVisible(true)}
              /> */}
            </Title>
            <EditCIProject
              data={data}
              visible={isModalVisible}
              onOk={onOk}
              onCancel={() => setIsModalVisible(false)}
            />
          </div>
          <div>
            <Descriptions
              className={styles.descriptions}
              size="small"
              column={2}
            >
              <Descriptions.Item label="团队">
                {data.team ? (
                  <a
                    href={`/projects/list?team_id=${data.team.id}`}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {data.team.name}
                  </a>
                ) : null}
              </Descriptions.Item>
              <Descriptions.Item label="语言">
                <a
                  href={`/projects/list?language=${data.language}`}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  {data.language}
                </a>
              </Descriptions.Item>
              <Descriptions.Item label="负责人">
                {join(map(data.owners, ({ name }) => name))}
              </Descriptions.Item>
              <Descriptions.Item
                label="收藏"
                style={{ display: 'flex', alignItems: 'center' }}
              >
                <StarFilled
                  onClick={() => handleColllectProject()}
                  style={{
                    fontSize: '18px',
                    cursor: 'pointer',
                    color: collect ? '#fadb14' : '',
                  }}
                />
              </Descriptions.Item>
            </Descriptions>
          </div>

          <p>{data.desc}</p>
          {(data?.labels ?? []).length > 0 &&
            data.labels.map((item) => (
              <Tag color="blue" key={item}>
                {(allLabels.find((i) => i.label === item) || {})?.name || null}
              </Tag>
            ))}
          <div className={styles.contentLink}>
            <a
              onClick={openGitLink}
              // href={`${data.dev_doc_url}`}
              // target='_blank' rel='noopener noreferer'
            >
              <GitlabOutlined /> Git地址
            </a>
            {data.api_doc_url ? (
              <a
                href={`${data.api_doc_url}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                <ApiOutlined /> 接口文档
              </a>
            ) : null}
            {data.dev_doc_url ? (
              <a
                href={`${data.dev_doc_url}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                <FileTextOutlined /> 开发文档
              </a>
            ) : null}
          </div>
        </div>
      )}
    </>
  );
};

export default ProjectInfo;

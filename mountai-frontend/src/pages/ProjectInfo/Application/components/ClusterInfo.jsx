import React, { useState, useEffect } from 'react';
import { Modal, Typography, Spin, message } from 'antd';
import ReactJson from 'react-json-view';
import { get } from 'lodash';
import { getApplicationDescribe } from '../service';

const { Paragraph } = Typography;

const ClusterInfo = (props) => {
  const { onClose, modalVisible, application, filterInfo, data } = props;

  const [describe, setDescribe] = useState();
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    (async () => {
      if (modalVisible) {
        if ((data?.running_status || []).length > 0) {
          const version = get(data, 'running_status[0].version', '');
          setLoading(true);

          try {
            const res = await getApplicationDescribe(version, {
              app_id: data.id,
              env_name: filterInfo.envname,
              cluster_name: data.cluster_name,
            });
            setDescribe(res);
          } catch (error) {
            message.error(error.message);
            setDescribe({});
          } finally {
            setLoading(false);
          }
        }
      }
    })();
  }, [modalVisible, application, filterInfo]);

  return (
    <Modal
      width={640}
      bodyStyle={{
        padding: '32px 40px 48px',
      }}
      destroyOnClose
      title={`集群详情${data.cluster_name}`}
      visible={modalVisible}
      okText={false}
      onOk={onClose}
      onCancel={onClose}
    >
      <Paragraph>
        <ul>
          {data?.slb_url ? (
            <li>
              <Paragraph>
                <a
                  href={`${data.slb_url}`}
                  rel="noopener noreferrer"
                  target="__blank"
                >
                  负载均衡
                </a>
              </Paragraph>
            </li>
          ) : (
            <></>
          )}

          {data?.sentry_project_url ? (
            <li key={data?.sentry_project_url}>
              sentry项目地址:
              <Paragraph copyable={{ text: `${data?.sentry_project_url}` }}>
                <a
                  href={`${data?.sentry_project_url}`}
                  rel="noopener noreferrer"
                  target="__blank"
                >
                  {data?.sentry_project_url}
                </a>
              </Paragraph>
            </li>
          ) : null}

          {data?.sentry_project_public_dsn ? (
            <li key={data?.sentry_project_public_dsn}>
              sentry dsn地址:
              <Paragraph
                copyable={{
                  text: `${data?.sentry_project_public_dsn}`,
                }}
              >
                <a
                  href={`${data?.sentry_project_public_dsn}`}
                  rel="noopener noreferrer"
                  target="__blank"
                >
                  {data?.sentry_project_public_dsn}
                </a>
              </Paragraph>
            </li>
          ) : null}

          {data?.access_hosts &&
            data.access_hosts.map((item) => (
              <li key={item}>
                {item.indexOf('.svc.cluster.local') > 0
                  ? '服务内网域名'
                  : '服务公网域名'}
                :
                <Paragraph copyable={{ text: `${item}` }}>
                  <a
                    href={`http://${item}`}
                    rel="noopener noreferrer"
                    target="__blank"
                  >
                    {item}
                  </a>
                </Paragraph>
              </li>
            ))}

          {(data?.kong_frontend_info || []).length > 0 && (
            <li>
              域名配置详情:
              <ReactJson
                src={data?.kong_frontend_info || []}
                collapsed
                displayDataTypes={false}
                displayObjectSize={false}
                name={null}
              />
            </li>
          )}

          {(data?.running_status || []).length > 0 && (
            <Spin spinning={loading} size="small">
              <li>
                状态详情:
                <ReactJson
                  src={describe}
                  collapsed
                  displayDataTypes={false}
                  displayObjectSize={false}
                  name={null}
                />
              </li>
            </Spin>
          )}
        </ul>
      </Paragraph>
    </Modal>
  );
};

export default ClusterInfo;

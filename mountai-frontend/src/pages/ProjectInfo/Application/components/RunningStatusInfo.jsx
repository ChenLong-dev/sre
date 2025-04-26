import React, { useState, useEffect } from 'react';
import { Modal, Spin, message } from 'antd';
import ReactJson from 'react-json-view';
import { getApplicationPodDescribe } from '../service';

const RunningStatusInfo = (props) => {
  const [loading, setLoading] = useState(false);
  const [podDescribe, setPodDescribe] = useState();

  const {
    onCancel: handleUpdateModalVisible,
    modalVisible,
    running,
    filterInfo,
    podData,
    clusterName,
    application,
    projectInfo,
  } = props;

  useEffect(() => {
    (async () => {
      if (modalVisible) {
        const version = running?.version;
        const podName = podData?.name;

        setLoading(true);
        try {
          const res = await getApplicationPodDescribe({
            version,
            podName,
            envName: filterInfo?.envname,
            clusterName,
            namespace: running.namespace,
            containerName: projectInfo.name + "-" + application.name,
          });
          setPodDescribe(res);
        } catch (error) {
          message.error(error.message);
          setPodDescribe({});
        } finally {
          setLoading(false);
        }
      }
    })();
  }, [modalVisible]);

  return (
    <Modal
      width={640}
      bodyStyle={{
        padding: '32px 40px 48px',
      }}
      destroyOnClose
      title="Pods详情"
      visible={modalVisible}
      onCancel={() => handleUpdateModalVisible()}
    >
      <h4>Pod 状态详情：</h4>
      <Spin spinning={loading}>
        <ReactJson
          src={podDescribe}
          displayDataTypes={false}
          displayObjectSize={false}
          name={null}
        />
      </Spin>
    </Modal>
  );
};

export default RunningStatusInfo;

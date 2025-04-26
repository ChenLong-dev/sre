import React, { useEffect, useRef, useState } from 'react';
import PropTypes from 'prop-types';
import { keyBy, map, set } from 'lodash';
import { message, Modal, Switch, Row, Col } from 'antd';
import { useEventCallback } from '@/lib/useEventCallback';
import { makeCancelablePromise } from '@/lib/makeCancelablePromise';
import { delay } from '@/lib/delay';
import { UPDATING, ENABLED } from '@/constants/clusterDNSStatus';
import { ACTIONS } from '@/utils/commonData';
import { getApplicationInfo, postTasks } from '../service';

async function getAppClusterDNSStatus({ appId, envName, clusterName }) {
  const app = await getApplicationInfo({
    appId,
    envName,
    clusterName,
  });

  return {
    status: app.in_cluster_dns_status,
    cluster: clusterName,
  };
}

function ClusterDNSStatus({ appId, clusters, env, visible, ...restProps }) {
  const [loading, setLoading] = useState(false);
  const [clusterDNSStatusMap, setClusterDNSStatusMap] = useState(null);

  const cancelablePromiseRef = useRef(null);

  const pollClusterDNSStatus = useEventCallback(async ({ clusterName }) => {
    cancelablePromiseRef.current?.cancel();
    let status = UPDATING;

    try {
      ({ status } = await getAppClusterDNSStatus({
        appId,
        envName: env,
        clusterName,
      }));
    } catch {
      //
    }

    if (status === UPDATING) {
      cancelablePromiseRef.current = makeCancelablePromise(delay(10000));
      await cancelablePromiseRef.current;
      return pollClusterDNSStatus({ clusterName });
    }

    setClusterDNSStatusMap({
      ...set(clusterDNSStatusMap, [clusterName, 'status'], status),
    });

    return status;
  });

  useEffect(() => {
    (async () => {
      if (visible) {
        setLoading(true);

        try {
          const clusterStatus = await Promise.all(
            map(clusters, ({ name }) => {
              return getAppClusterDNSStatus({
                appId,
                envName: env,
                clusterName: name,
              });
            })
          );

          setClusterDNSStatusMap(keyBy(clusterStatus, 'cluster'));
        } finally {
          setLoading(false);
        }
      }
    })();

    return () => cancelablePromiseRef.current?.cancel();
  }, [appId, env, visible]);

  const updateClusterDNSStatus = useEventCallback(
    async ({ clusterName, action }) => {
      setLoading(true);
      try {
        await postTasks({
          clusterName,
          appId,
          envName: env,
          action,
          param: {},
        });
        await pollClusterDNSStatus({ clusterName });
      } catch (error) {
        message.error(error.message);
      } finally {
        setLoading(false);
      }
    }
  );

  return (
    <Modal
      title="集群DNS"
      footer={null}
      visible={visible}
      width={200}
      {...restProps}
    >
      <Row gutter={[16, 24]}>
        {map(clusters, ({ name }) => (
          <React.Fragment key={name}>
            <Col span={10} style={{ textAlign: 'right' }}>{`${name}:`}</Col>
            <Col span={14}>
              <Switch
                loading={loading}
                checked={clusterDNSStatusMap?.[name]?.status === ENABLED}
                onChange={(checked) =>
                  updateClusterDNSStatus({
                    clusterName: name,
                    action: checked
                      ? ACTIONS.ENABLE_IN_CLUSTER_DNS
                      : ACTIONS.DISABLE_IN_CLUSTER_DNS,
                  })
                }
              />
            </Col>
          </React.Fragment>
        ))}
      </Row>
    </Modal>
  );
}

ClusterDNSStatus.propTypes = {
  appId: PropTypes.string.isRequired,
  env: PropTypes.string.isRequired,
  clusters: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
    })
  ).isRequired,
  visible: PropTypes.bool.isRequired,
};

export default React.memo(ClusterDNSStatus);

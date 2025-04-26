import React, { useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { keyBy, map, mapValues, reduce, toNumber } from 'lodash';
import { message, Modal, Form, InputNumber, Alert } from 'antd';
import { useEventCallback } from '@/lib/useEventCallback';
import { postAppClusterWeights, getAppClusterWeights } from '../service';

function ClusterWeights({ appId, clusters, env, onOk, visible, ...restProps }) {
  const [form] = Form.useForm();
  const [confirmLoading, setConfirmLoading] = useState(false);

  useEffect(() => {
    (async () => {
      if (visible) {
        const { list: weights } = await getAppClusterWeights({ appId, env });

        const values = mapValues(keyBy(weights, 'cluster_name'), 'weight');

        form.setFieldsValue(values);
      }
    })();
  }, [appId, env, visible]);

  const updateAppClusterWeights = useEventCallback(
    async ({ clusterWeights, forceUpdateAll = false }) => {
      setConfirmLoading(true);
      try {
        await postAppClusterWeights({
          appId,
          env,
          clusterWeights,
          forceUpdateAll,
        });
        setConfirmLoading(false);
        onOk?.();
      } catch (error) {
        if (error.code === 9010066) {
          Modal.confirm({
            title: '确认强制将所有人工配置的权重全部统一为目标值？',
            onOk: () =>
              updateAppClusterWeights({ clusterWeights, forceUpdateAll: true }),
            onCancel: () => setConfirmLoading(false),
          });
        } else {
          message.error(error.message);
          setConfirmLoading(false);
        }
      }
    }
  );

  const onSubmit = useEventCallback(async () => {
    const values = await form.validateFields();

    const total = reduce(values, (p, c) => p + c, 0);

    if (total !== 100) {
      message.error('集群权重总和必须等于100');
      return null;
    }

    const clusterWeights = map(values, (weight, clusterName) => ({
      cluster_name: clusterName,
      weight,
    }));
    return updateAppClusterWeights({ appId, env, clusterWeights });
  });

  return (
    <Modal
      title="集群权重"
      onOk={onSubmit}
      confirmLoading={confirmLoading}
      visible={visible}
      {...restProps}
    >
      {clusters.length > 1 ? (
        <Form form={form} wrapperCol={{ span: 8 }} labelCol={{ span: 6 }}>
          {map(clusters, ({ name }) => (
            <Form.Item
              key={name}
              name={name}
              label={`${name}权重`}
              initialValue={0}
              rules={[
                {
                  required: true,
                  message: '集群权重必填',
                },
                {
                  pattern: /^\d+$/,
                  message: '集群权重必须为正整数',
                },
                {
                  validator: async (_, value) => {
                    const weight = toNumber(value);
                    if (weight > 100 || weight < 0) {
                      throw new Error('集群权重应在0到100之间');
                    }
                  },
                },
              ]}
            >
              <InputNumber />
            </Form.Item>
          ))}
        </Form>
      ) : (
        <Alert showIcon type="warning" message="当前应用未部署到多个集群" />
      )}
    </Modal>
  );
}

ClusterWeights.propTypes = {
  appId: PropTypes.string.isRequired,
  env: PropTypes.string.isRequired,
  clusters: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
    })
  ),
  visible: PropTypes.bool.isRequired,
  onOk: PropTypes.func,
};

ClusterWeights.defaultProps = {
  onOk: null,
  clusters: [],
};

export default React.memo(ClusterWeights);

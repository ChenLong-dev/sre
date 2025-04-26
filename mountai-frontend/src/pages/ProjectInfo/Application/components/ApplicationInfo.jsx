import React, { useState, useCallback, useMemo } from 'react';
import { Checkbox, Modal, Typography } from 'antd';
import { map } from 'lodash';
import { postUserSubscriptions, deleteUserSubscriptions } from '../service';
import { SUBSCRIPTION_OPTIONS } from '../constants/subscriptionOptions';

const { Paragraph } = Typography;

const ApplicationInfo = (props) => {
  const { onClose, modalVisible, filterInfo, data } = props;

  const [subscriptions, setSubscriptions] = useState();

  const checkedSubscriptions = useMemo(
    () => subscriptions || data.subscriptions,
    [subscriptions, data]
  );

  const onSubscriptionsChange = useCallback(
    (e) => {
      if (e.target.checked) {
        postUserSubscriptions({
          app_id: data.id,
          env_name: filterInfo.envname,
          action: e.target.value,
        });
      } else {
        deleteUserSubscriptions({
          app_id: data.id,
          env_name: filterInfo.envname,
          action: e.target.value,
        });
      }
    },
    [data, filterInfo]
  );

  return (
    <Modal
      width={640}
      bodyStyle={{
        padding: '32px 40px 48px',
      }}
      destroyOnClose
      title={`应用详情${data.name ? ` - ${data.name}` : ''}`}
      visible={modalVisible}
      okText={false}
      onOk={onClose}
      onCancel={onClose}
    >
      <Paragraph>
        <ul>
          <li>
            应用id:
            <Paragraph copyable>{data.id}</Paragraph>
          </li>
          <li>
            订阅应用事件:
            <Checkbox.Group
              style={{ width: '100%' }}
              value={checkedSubscriptions}
              onChange={(checkedValues) => setSubscriptions(checkedValues)}
            >
              {map(SUBSCRIPTION_OPTIONS, (event) => (
                <Checkbox
                  key={event}
                  value={event}
                  onChange={onSubscriptionsChange}
                >
                  {event}
                </Checkbox>
              ))}
            </Checkbox.Group>
          </li>
        </ul>
      </Paragraph>
    </Modal>
  );
};

export default ApplicationInfo;

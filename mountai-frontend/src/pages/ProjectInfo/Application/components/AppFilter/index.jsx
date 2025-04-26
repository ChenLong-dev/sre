import React, { useState, useCallback, useMemo } from 'react';
import { useLocation, useHistory } from 'umi';
import {
  Button,
  Space,
  Radio,
  Input,
  Row,
  Col,
  message,
  Form,
  Select,
} from 'antd';
import { map } from 'lodash';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import {
  SERVICE,
  WORKER,
  CRON_JOB,
  ONE_TIME_JOB,
} from '../../../../../constants/applicationTypes';
import styles from './index.less';
import YamlCopyList from './YamlCopyList';
import { getProjectConfig } from '../../service';
import { hasDeveloperPermission } from '../../utils/hasPermission';

const appTypes = [SERVICE, WORKER, CRON_JOB, ONE_TIME_JOB];

const AppFilter = ({
  createApp,
  filterInfo,
  setFilterInfo,
  setReloadingApplications,
  projectId,
  batchAction,
  memberRole,
}) => {
  const [form] = Form.useForm();
  const location = useLocation();
  const history = useHistory();
  const [showYamlVisible, setShowYamlVisible] = useState(false);
  const [yamlData, setYamlData] = useState({});

  const setAppEenChange = (e) => {
    let newFilter = { envname: e.target.value };
    if (filterInfo.keyword) {
      newFilter = {
        ...newFilter,
        keyword: filterInfo.keyword,
      };
    }
    setFilterInfo(newFilter);
    // 设置路由
    if (JSON.stringify(newFilter) !== '{}') {
      const url = new URLSearchParams(newFilter);
      window.history.replaceState(null, null, `?${url.toString()}`);
    }

    setReloadingApplications(true);
  };

  const openProjectConfig = async () => {
    try {
      const result = await getProjectConfig(projectId, filterInfo.envname);
      setYamlData(result);
      setShowYamlVisible(true);
    } catch (error) {
      message.error(error.message || '没找到配置');
    }
  };

  const search = useCallback(() => {
    const values = form.getFieldsValue();
    history.push(
      `${location.pathname}?keyword=${values.keyword ?? ''}&type=${
        values.type ?? ''
      }`
    );
  }, [form]);

  const initialValues = useMemo(() => {
    const { keyword, type } = location.query ?? {};
    return {
      keyword,
      type,
    };
  }, []);

  return (
    <div className={styles.appFilter}>
      <div className={styles.label}>
        <span>应用列表</span>
      </div>
      <div className={styles.content}>
        <Row style={{ marginBottom: '15px' }}>
          <Col span={21}>
            <Radio.Group
              className="analysis-relative-time"
              onChange={setAppEenChange}
              defaultValue={filterInfo.envname}
              value={filterInfo.envname}
            >
              <Radio.Button value="test">TEST</Radio.Button>
              {/* <Radio.Button value="fat">FAT</Radio.Button> */}
              <Radio.Button value="prod">PROD</Radio.Button>
            </Radio.Group>
            {/* {hasDeveloperPermission(memberRole, filterInfo.envname) && (
              <Button
                type="primary"
                style={{ marginLeft: '20px' }}
                onClick={batchAction}
              >
                批量操作
              </Button>
            )} */}
          </Col>

          <Col span={3}>
            {hasDeveloperPermission(memberRole, filterInfo.envname) && (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => createApp(true)}
              >
                添加应用
              </Button>
            )}
          </Col>
        </Row>

        <Row style={{ marginBottom: '5px' }}>
          <Col span={10}>
            <Space>
              {/* {hasDeveloperPermission(memberRole, filterInfo.envname) && (
                <Button onClick={() => openProjectConfig()}> 最新配置</Button>
              )} */}
              <Button
                icon={<ReloadOutlined />}
                onClick={() => setReloadingApplications(true)}
              >
                刷新
              </Button>
            </Space>
          </Col>
          <Col span={14}>
            <Form layout="inline" form={form} initialValues={initialValues}>
              <Form.Item label="应用名称" name="keyword">
                <Input allowClear placeholder="请输入应用名称" />
              </Form.Item>
              <Form.Item label="应用类型" name="type">
                <Select allowClear style={{ width: 100 }}>
                  {map(appTypes, (type) => (
                    <Select.Option key={type} value={type}>
                      {type}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
              <Form.Item>
                <Button type="primary" onClick={search}>
                  查询
                </Button>
              </Form.Item>
            </Form>
          </Col>
        </Row>
      </div>
      {showYamlVisible ? (
        <YamlCopyList
          filterInfo={filterInfo}
          yamlData={yamlData}
          showYamlVisible={showYamlVisible}
          handleSetVisible={(bol) => setShowYamlVisible(bol)}
        />
      ) : null}
    </div>
  );
};

export default AppFilter;

import React, {
  useState,
  useRef,
  useEffect,
  useCallback,
  useMemo,
} from 'react';
import {
  Button,
  Space,
  Menu,
  Divider,
  Row,
  Col,
  Modal,
  Dropdown,
  message,
  Tooltip,
  Form,
  Switch,
  Radio,
} from 'antd';
import {
  FormOutlined,
  DownOutlined,
  UpOutlined,
  ExclamationCircleOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import querystring from 'querystring';
import { filter, map, find } from 'lodash';
import { connect } from 'umi';
import { useEventCallback } from '@/lib/useEventCallback';
import { makeCancelablePromise } from '@/lib/makeCancelablePromise';
import { RESTFUL } from '@/constants/applicationServiceTypes';
import { INGRESS } from '@/constants/applicationServiceExposeTypes';
import CreateTasks from '../CreateTasks';
import NeedCorrect from '../NeedCorrect';
import ApplicationInfo from '../ApplicationInfo';
import SettingApplication from '../SettingApplication';
import RunningStatusInfo from '../RunningStatusInfo';
import {
  postTasks,
  getApplicationInfo,
  postAppCorrectName,
  deleteApp,
  updateApp,
  getProjectimageTag,
  postSentry,
  getApplicationTips,
  getAppClusters,
} from '../../service';
import RunItem from '../RunItem';
import {
  hasDeveloperPermission,
  hasMaintainerPermission,
} from '../../utils/hasPermission';
import ClusterInfo from '../ClusterInfo';
import ClusterWeights from '../ClusterWeights';
import ClusterDNSStatus from '../ClusterDNSStatus';
import styles from './index.less';

const { confirm } = Modal;

const ServerNamerender = ({ type }) => {
  if (type === 'CronJob') {
    return (
      <div
        className={styles.appMark}
        style={{ backgroundColor: 'rgb(5, 191, 39)' }}
      >
        C
      </div>
    );
  }
  if (type === 'OneTimeJob') {
    return (
      <div
        className={styles.appMark}
        style={{ backgroundColor: 'rgb(10,101,89)' }}
      >
        O
      </div>
    );
  }
  if (type === 'Worker') {
    return (
      <div
        className={styles.appMark}
        style={{ backgroundColor: 'rgb(53, 207, 201)' }}
      >
        W
      </div>
    );
  }

  return <div className={styles.appMark}>S</div>;
};

const AppContainer = (props) => {
  const [form] = Form.useForm();

  const {
    application,
    filterInfo,
    setReloadingApplications,
    projectId,
    projectInfo,
    memberRole,
    handleCreatePprof,
    nodeLabels,
    projectClusterMap,
  } = props;

  const timer = useRef();
  const applicationInfoPromise = useRef();
  const [createTasksModal, setCreateTasksModal] = useState(false);
  const [needCorrectModal, setNeedCorrectModal] = useState(false);
  const [applicationInfoModal, setApplicationInfoModal] = useState(false);
  const [clusterInfoModal, setClusterInfoModal] = useState(false);
  const [settingApplication, setSettingApplicationModal] = useState(false);
  const [applicationInfo, setApplicationInfo] = useState({});
  const [applicationTips, setApplicationTips] = useState({}); // 应用的提示信息
  const [currentClusterName, setCurrentClusterName] = useState();
  const [newImageTags, setNewImageTags] = useState([]);
  const [showAllPods, setShowAllPods] = useState(false); // 是否展开查看所有定时任务
  const [runningStatusInfoModal, setRunningStatusInfoModal] = useState(false); // pod describe详情弹窗
  const [actPodData, setActPodData] = useState({}); // 当前点击的pod信息
  const [actRunning, setActRunning] = useState({}); // 当前点击pod状态所属的running

  const [appClusters, setAppClusters] = useState();
  const [clusterWeights, setClusterWeights] = useState(false);
  const [clusterDNSStatus, setClusterDNSStatus] = useState(false);

  const showClusterWeights = useEventCallback(() => setClusterWeights(true));

  const hideClusterWeights = useEventCallback(() => setClusterWeights(false));

  const showClusterDNSStatus = useEventCallback(() =>
    setClusterDNSStatus(true)
  );

  const hideClusterDNSStatus = useEventCallback(() =>
    setClusterDNSStatus(false)
  );

  const projectClusters = useMemo(
    () =>
      filter(
        projectClusterMap[projectId],
        ({ env }) => env === filterInfo.envname
      ),
    [projectId, projectClusterMap, filterInfo.envname]
  );

  const defaultCluster = useMemo(() => find(projectClusters, 'is_default'), [
    projectClusters,
  ]);

  const updateApplicationInfo = useCallback(async () => {
    if (currentClusterName) {
      applicationInfoPromise.current?.cancel();
      applicationInfoPromise.current = makeCancelablePromise(
        getApplicationInfo({
          appId: application.id,
          envName: filterInfo.envname,
          clusterName: currentClusterName,
        })
      );
      const newApplicationInfo = await applicationInfoPromise.current;
      setApplicationInfo(newApplicationInfo);
      return newApplicationInfo;
    }
    return undefined;
  }, [application, filterInfo, currentClusterName, defaultCluster]);

  const updateApplicationTips = useCallback(async () => {
    if (currentClusterName) {
      let newApplicationTips = {};

      if (application?.id && filterInfo?.envname) {
        newApplicationTips = await getApplicationTips(application.id, filterInfo?.envname, currentClusterName);
        setApplicationTips(newApplicationTips || {});
      }

      return newApplicationTips;
    }
    return undefined;
  }, [application, filterInfo, currentClusterName]);

  const timeloop = useEventCallback(() => {
    updateApplicationInfo().then((res) => {
      if (timer.current) {
        clearTimeout(timer.current);
        timer.current = null;
      }

      if (
        res?.running_status.every((i) =>
          ['success', 'resume'].includes(i.task_display_icon)
        )
      ) {
        return;
      }
      timer.current = setTimeout(() => {
        timeloop();
      }, 5000);
    });
  });

  const startDeploy = async () => {
    if (applicationInfo && applicationInfo.need_correct) {
      // 弹出迁移弹出层先迁移

      setNeedCorrectModal(true);
    } else {
      const response = await getProjectimageTag(projectId);
      setNewImageTags(response.list);
      setCreateTasksModal(true);
    }
  };

  const updateAppClusters = useCallback(async () => {
    const { list } = await getAppClusters(application.id, {
      env_name: filterInfo.envname,
    });
    setAppClusters(list);
  }, [application.id, filterInfo.envname]);

  /**
   * 提交创建发布任务
   * @param {*} values
   */
  const createNewTask = async ({ cluster_name, ...values }, bol) => {
    const postdata = {
      clusterName: cluster_name,
      appId: application.id,
      envName: filterInfo.envname,
    };
    if (bol) postdata.ignoreExpectedBranch = true;
    if (
      filterInfo.envname === 'prod' &&
      !['CronJob', 'OneTimeJob'].includes(application.type) && !values.disable_canary
    ) {
      postdata.action = 'canary_deploy';
    } else {
      postdata.action = 'full_deploy';
    }

    const { description, ...param } = values;
    postdata.param = param;
    postdata.description = description;

    try {
      await postTasks(postdata);
      message.success('任务创建成功');
      setCreateTasksModal(false);
      updateAppClusters();
      setCurrentClusterName(cluster_name);
      // 轮询
      timeloop();
    } catch (error) {
      if (error.code === 9010036) {
        Modal.confirm({
          title: '提示',
          content:
            '检测到当前该环境部署的分支和之前的发布分支不一致，是否确认发布？',
          cancelText: '取消',
          okText: '确认',
          onOk: () => {
            createNewTask({ cluster_name, ...values }, true);
          },
          onCancel: () => setCreateTasksModal(false),
        });
      } else {
        message.error(error.message || '任务创建失败');
      }
    }
  };

  // 更新应用
  const postCorrect = async (values) => {
    try {
      await postAppCorrectName(values);
      message.success('修改成功');
      setNeedCorrectModal(false);
      setReloadingApplications(true);
    } catch (error) {
      message.error(error.message || '修改失败');
    }
  };

  const showApplicationInfo = () => {
    setApplicationInfoModal(true);
  };

  const deleteApplication = async () => {
    const value = form.getFieldsValue();
    try {
      await deleteApp(application.id, value);
      message.success('删除成功');
      setNeedCorrectModal(false);
      setReloadingApplications(true);
    } catch (error) {
      message.error(error.message || '删除失败');
    }
  };

  const deleteConfirm = () => {
    confirm({
      title: `确定删除${application.name && application.name}吗?`,
      icon: <ExclamationCircleOutlined />,
      content: applicationInfo.sentry_project_public_dsn && (
        <Form form={form}>
          <Form.Item
            name="delete_sentry"
            label="是否同时删除sentry项目"
            valuePropName="checked"
            initialValue={false}
          >
            <Switch />
          </Form.Item>
        </Form>
      ),
      onOk() {
        deleteApplication();
      },
    });
  };

  const postApplicationSentry = () => {
    // 为应用添加sentry
    Modal.confirm({
      title: `提示${application.name && ` - ${application.name}`}`,
      content: '确认一键创建sentry项目？',
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        try {
          await postSentry(application.id);
          message.success('设置sentry成功');
        } catch (error) {
          message.error(error.message || '操作失败');
        }
      },
      onCancel: () => null,
    });
  };

  const onChangeCluster = useEventCallback((e) => {
    setCurrentClusterName(e.target.value);
  });

  const menu = (
    <Menu>
      <Menu.Item>
        <a onClick={showApplicationInfo}>详情</a>
      </Menu.Item>

      {/* {applicationInfo?.service_type === RESTFUL && (
        <Menu.Item>
          <a onClick={showClusterWeights}>集群权重</a>
        </Menu.Item>
      )} */}

      {/* {applicationInfo?.service_type === RESTFUL &&
        applicationInfo?.service_expose_type === INGRESS && (
          <Menu.Item>
            <a onClick={showClusterDNSStatus}>集群DNS</a>
          </Menu.Item>
        )} */}

      {/* {!applicationInfo?.sentry_project_public_dsn && (
        <Menu.Item>
          <a onClick={postApplicationSentry}>sentry</a>
        </Menu.Item>
      )} */}

      {hasDeveloperPermission(memberRole, filterInfo.envname) && (
        <Menu.Item>
          <a onClick={() => setSettingApplicationModal(true)}>修改</a>
        </Menu.Item>
      )}

      {hasMaintainerPermission(memberRole, filterInfo.envname) && (
        <Menu.Item>
          <a onClick={deleteConfirm}>删除</a>
        </Menu.Item>
      )}
    </Menu>
  );

  const logMenu = (
    <Menu>
      {applicationInfo && applicationInfo.log_store_url_based_project && (
        <Menu.Item>
          <Tooltip
            title={`检索时请使用形如“${
              querystring.parse(
                applicationInfo.log_store_url_based_project.split('?')[1] || ''
              )?.queryString || ''
            }”的检索条件进行精确检索，本链接已提供该检索条件。`}
          >
            <a
              href={applicationInfo.log_store_url_based_project}
              target="_blank"
              rel="noopener noreferrer"
            >
              云日志
            </a>
          </Tooltip>
        </Menu.Item>
      )}
    </Menu>
  );

  const applicationWasted =
    applicationTips.wasted_max_cpu_usage_rate ||
    applicationTips.wasted_max_mem_usage_rate;

  useEffect(() => {
    updateApplicationTips();
  }, [updateApplicationTips]);

  useEffect(() => {
    (async () => {
      try {
        const newApplicationInfo = await updateApplicationInfo();
        if (
          newApplicationInfo?.running_status &&
          !newApplicationInfo.running_status.every(
            (i) => i.task_display_icon === 'success'
          )
        ) {
          timeloop();
        }
      } catch (error) {
        if (!error.canceled) {
          message.error(error.message);
        }
      }
    })();
  }, [updateApplicationInfo]);

  useEffect(() => {
    updateAppClusters();
  }, [updateAppClusters]);

  useEffect(() => {
    if (!currentClusterName && appClusters) {
      setCurrentClusterName(appClusters[0]?.name || defaultCluster?.name);
    }
  }, [appClusters, defaultCluster]);

  return (
    <div className={styles.appContainer}>
      <Row style={{ flexFlow: 'row' }} className={styles.appContainerHover}>
        <Col span={6} className={styles.leftContainer}>
          <div className={styles.appInfo}>
            <ServerNamerender type={application.type} />
            <div className={styles.appWork}>
              <Space size={4} className={styles.appBasic}>
                {hasDeveloperPermission(memberRole, filterInfo.envname) && (
                  <FormOutlined
                    onClick={() => setNeedCorrectModal(true)}
                    style={{ color: '#1890ff' }}
                  />
                )}

                <div
                  title={application.name}
                  className={`${styles.appName} ${
                    applicationWasted ? styles.wasted : ''
                  }`}
                >
                  {application.name}
                </div>

                {applicationWasted ? (
                  <Tooltip
                    placement="topLeft"
                    title={
                      <>
                        {applicationTips.wasted_max_cpu_usage_rate && (
                          <div>
                            当前cpu处于浪费状态，7天内最大cpu使用率：
                            {applicationTips.wasted_max_cpu_usage_rate}
                          </div>
                        )}
                        {applicationTips.wasted_max_mem_usage_rate && (
                          <div>
                            当前内存处于浪费状态，7天内最大内存使用率：
                            {applicationTips.wasted_max_mem_usage_rate}
                          </div>
                        )}
                      </>
                    }
                  >
                    <ExclamationCircleOutlined className={styles.wastedMark} />
                  </Tooltip>
                ) : null}

                {application.description && (
                  <Tooltip
                    placement="topLeft"
                    title={
                      <div style={{ whiteSpace: 'pre-line' }}>
                        {application.description}
                      </div>
                    }
                  >
                    <QuestionCircleOutlined />
                  </Tooltip>
                )}
              </Space>
              <Space align="center" className={styles.appworkClick}>
                {hasDeveloperPermission(memberRole, filterInfo.envname) && (
                  <Button
                    size="small"
                    type="primary"
                    className={styles.appWorkButtion}
                    onClick={() => startDeploy()}
                  >
                    发布
                  </Button>
                )}
                <Dropdown overlay={logMenu}>
                  <a
                    style={{ whiteSpace: 'nowrap' }}
                    className="ant-dropdown-link"
                    onClick={(e) => e.preventDefault()}
                  >
                    云日志 <DownOutlined />
                  </a>
                </Dropdown>

                <Dropdown overlay={menu}>
                  <a
                    style={{ whiteSpace: 'nowrap' }}
                    className="ant-dropdown-link"
                    onClick={(e) => e.preventDefault()}
                  >
                    更多
                    <DownOutlined />
                  </a>
                </Dropdown>
              </Space>
            </div>
          </div>
          <Divider type="vertical" className={styles.appInfoDivider} />
        </Col>
        <Col span={19} className={styles.runningList}>
          <div
            className={
              ['CronJob', 'Worker', 'Service', 'OneTimeJob'].includes(applicationInfo?.type)
                ? `${styles.defaultRunningPods} ${
                    showAllPods &&
                    (applicationInfo.running_status.length > 0 ||
                      appClusters?.length > 1)
                      ? ''
                      : styles.noRunningPods
                  }`
                : ''
            }
          >
            {['CronJob', 'Worker', 'Service'].includes(applicationInfo?.type) &&
            (applicationInfo.running_status.length > 0 ||
              appClusters?.length > 1) ? (
              <div
                className={`${showAllPods ? styles.viewLess : styles.viewMore}`}
                onClick={() => setShowAllPods(!showAllPods)}
              >
                {showAllPods ? '折叠' : '展开'}列表&nbsp;
                {showAllPods ? <UpOutlined /> : <DownOutlined />}
              </div>
            ) : null}
            <Col span={5}>
              <Radio.Group
                value={currentClusterName}
                onChange={onChangeCluster}
                size="small"
              >
                <Space direction="vertical">
                  {map(appClusters, ({ name }) => (
                    <Radio.Button key={name} value={name}>
                      {name}
                    </Radio.Button>
                  ))}
                </Space>
              </Radio.Group>
              <div style={{ marginTop: 5 }}>
                <Space>
                  <Button
                    type="primary"
                    size="small"
                    onClick={() => {
                      setClusterInfoModal(true);
                    }}
                  >
                    详情
                  </Button>
                  {applicationInfo?.monitor_url && (
                    <Button
                      type="primary"
                      size="small"
                      href={applicationInfo ? applicationInfo?.monitor_url : ''}
                      target="_blank"
                    >
                      监控
                    </Button>
                  )}
                </Space>
              </div>
            </Col>
            <Col span={19}>
              {applicationInfo?.running_status &&
                applicationInfo.running_status.length > 0 &&
                applicationInfo.running_status.map((item) => {
                  return (
                    <RunItem
                      key={JSON.stringify(item)}
                      running={item}
                      autoLoadApplication={timeloop}
                      handleStopTaskImage={timeloop}
                      handleCreatePprof={(pod, running, applicationData) =>
                        handleCreatePprof(pod, running, {
                          ...applicationData,
                          cluster_name: currentClusterName,
                        })
                      }
                      handleChangeVisible={(visible) =>
                        setRunningStatusInfoModal(visible)
                      }
                      handleChangeActPod={(pod) => setActPodData(pod)}
                      handleChangeActRunning={(running) =>
                        setActRunning(running)
                      }
                      application={application}
                      filterInfo={filterInfo}
                      projectInfo={projectInfo}
                      memberRole={memberRole}
                      clusterName={currentClusterName}
                    />
                  );
                })}
            </Col>
          </div>
        </Col>
      </Row>

      <CreateTasks
        onSubmit={createNewTask}
        onCancel={() => {
          setCreateTasksModal(false);
          // 可能要多做一点事
        }}
        nodeLabels={nodeLabels}
        modalVisible={createTasksModal}
        application={application}
        applicationInfo={applicationInfo}
        projectInfo={projectInfo}
        filterInfo={filterInfo}
        imageTags={newImageTags}
        applicationTips={applicationTips}
        projectClusters={projectClusters}
        defaultClusterName={defaultCluster?.name}
        clusterName={currentClusterName}
      />

      <NeedCorrect
        onSubmit={postCorrect}
        onCancel={() => {
          setNeedCorrectModal(false);
          // 可能要多做一点事
        }}
        modalVisible={needCorrectModal}
        applicationType={application.type}
        application={application}
        filterInfo={filterInfo}
      />

      <ApplicationInfo
        onClose={() => {
          setApplicationInfoModal(false);
        }}
        modalVisible={applicationInfoModal}
        application={application}
        filterInfo={filterInfo}
        data={applicationInfo}
      />

      <ClusterInfo
        onClose={() => {
          setClusterInfoModal(false);
        }}
        modalVisible={clusterInfoModal}
        application={application}
        filterInfo={filterInfo}
        data={applicationInfo}
      />

      <SettingApplication
        onSubmit={updateApp}
        onReload={() => setReloadingApplications(true)}
        onCancel={() => {
          setSettingApplicationModal(false);
        }}
        modalVisible={settingApplication}
        applicationType={application.type}
        applicationInfo={applicationInfo}
        filterInfo={filterInfo}
      />

      <RunningStatusInfo
        onCancel={() => {
          setRunningStatusInfoModal(false);
        }}
        modalVisible={runningStatusInfoModal}
        running={actRunning}
        filterInfo={filterInfo}
        podData={actPodData}
        clusterName={currentClusterName}
        application={application}
        projectInfo={projectInfo}
      />

      <ClusterWeights
        destroyOnClose
        visible={clusterWeights}
        onCancel={hideClusterWeights}
        onOk={hideClusterWeights}
        appId={application?.id}
        env={filterInfo.envname}
        clusters={appClusters}
      />

      <ClusterDNSStatus
        destroyOnClose
        visible={clusterDNSStatus}
        onCancel={hideClusterDNSStatus}
        appId={application?.id}
        env={filterInfo.envname}
        clusters={projectClusters}
      />

      <Divider type="horizontal" className={styles.appContainerDivider} />
    </div>
  );
};

export default connect((state) => ({
  projectClusterMap: state.application.projectClusterMap,
}))(AppContainer);

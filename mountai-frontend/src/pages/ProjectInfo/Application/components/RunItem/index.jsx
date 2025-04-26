import {
  Button,
  Space,
  Spin,
  Divider,
  Row,
  Col,
  Badge,
  Modal,
  Dropdown,
  Menu,
  message,
  Tooltip,
  Descriptions,
  Form,
  Input,
  Select,
} from 'antd';
import React, { useState, useEffect, useCallback } from 'react';
import {
  CloseCircleFilled,
  DownOutlined,
  ExclamationCircleOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import StatusIcon from '@/components/StatusIcon';
import InfoModal from '@/components/InfoModal';
import { find, includes, map } from 'lodash';
import {
  CRON_JOB,
  ONE_TIME_JOB,
  SERVICE,
  WORKER,
} from '@/constants/applicationTypes';
import {
  postTasks,
  getRunningtask,
  getRunningStatus,
  getRunningPodLog,
  getLatestTasks,
  setImageStatusStop,
  getApplicationInfo,
  deleteJob,
} from '../../service';
import styles from './index.less';
import TaskParamsDetail from '../TaskParamsDetail';
import UpdateHpaModal from '../UpdateHpaModal';
import {
  hasMaintainerPermission,
  hasDeveloperPermission,
} from '../../utils/hasPermission';

const { confirm } = Modal;

const { Option } = Select;

const RunningStatus = ({ status, task_display_icon }) => {
  let currentStatus = status;
  if (
    task_display_icon === 'underway' ||
    task_display_icon === 'canary_continue'
  ) {
    currentStatus = 'LOADING';
  }
  if (status === 'success') {
    currentStatus = 'SUCCESS';
  }
  return <StatusIcon status={currentStatus} />;
};

const PodsStatus = ({ status }) => {
  let currentStatus = status;
  if (status === 'success' || status === 'Running') {
    currentStatus = 'SUCCESS';
  }

  if (status === 'Pending') {
    currentStatus = 'LOADING';
  }

  return <StatusIcon status={currentStatus} />;
};

const JobsPodStatus = ({ status }) => {
  let currentStatus = status;
  if (status === 'Succeeded') {
    currentStatus = 'SUCCESS';
  }

  if (status === 'Running' || status === 'Pending') {
    currentStatus = 'LOADING';
  }

  return <StatusIcon status={currentStatus} />;
};

const RunItem = ({
  autoLoadApplication,
  running,
  clusterName,
  projectInfo,
  application,
  filterInfo,
  memberRole,
  handleStopTaskImage,
  handleChangeVisible,
  handleCreatePprof,
  handleChangeActPod,
  handleChangeActRunning,
}) => {
  const [form] = Form.useForm();

  const [runningData, setRunningData] = useState({});
  const [runngingPods, setRunngingPods] = useState([]);
  const [runngingJobs, setRunngingJobs] = useState([]);
  const [taskDetailVisible, setTaskDetailVisible] = useState(false);
  const [updateHpaVisible, setUpdateHpaVisible] = useState(false);

  const [loading, setLoading] = useState(false);

  const getRunningStatusData = () => {
    setLoading(true);
    getRunningStatus({
      clusterName,
      version: running.version,
      envName: filterInfo.envname,
      appId: application.id,
      namespace: running.namespace,
    }).then((res) => {
      if (res) {
        setLoading(false);
        setRunningData(res);
        if (res.deployment_pods) {
          setRunngingPods(res.deployment_pods);
        }
        if (res.jobs) {
          setRunngingJobs(res.jobs);
        }
      }
    });
  };

  useEffect(() => {
    if (running.version && filterInfo.envname && application.id) {
      getRunningStatusData();
    }
  }, [running]);

  const openUpdateHpaModal = () => {
    setUpdateHpaVisible(true);
  };

  /**
   *  打开运行日志
   */
  const openRunningInfoLog = async () => {
    const result = await getRunningtask({
      taskId: running.task_id,
    });
    if (result.detail && typeof result.detail === 'string') {
      InfoModal({
        content: result.detail,
        title: `日志${application.name && `${` - ${application.name}`}`}`,
      });
    } else {
      message.error('没找到...');
    }
  };

  /**
   *  打开详情
   */
  const openCronJobInfo = async () => {
    try {
      const result = await getApplicationInfo({
        appId: application.id,
        envName: filterInfo.envname,
        clusterName,
      });
      const runningStatus = find(result.running_status, [
        'version',
        running.version,
      ]);
      InfoModal({
        content: (
          <Descriptions column={1}>
            <Descriptions.Item label="上次调度时间">
              {runningStatus.last_schedule_time}
            </Descriptions.Item>
            <Descriptions.Item label="下次调度时间">
              {runningStatus.next_schedule_time}
            </Descriptions.Item>
          </Descriptions>
        ),
        title: '详情',
      });
    } catch (error) {
      message.error(error.message);
    }
  };

  const onUpdateHpaSubmit = useCallback(
    async (values) => {
      try {
        await postTasks({
          clusterName,
          appId: application.id,
          envName: filterInfo.envname,
          action: 'update_hpa',
          version: running.version,
          param: values,
          namespace: running.namespace,
        });
        message.success('配置完成');
        setUpdateHpaVisible(false);
      } catch (error) {
        message.error(error.message || '配置失败');
      }
    },
    [application, filterInfo, running]
  );

  //  重启运行实例
  const restartRunning = async () => {
    const values = await form.validateFields();

    const { setting, config_commit_id: configCommitId, ...param } = values;

    const postdata = {
      clusterName,
      appId: application.id,
      envName: filterInfo.envname,
      action: 'restart',
      version: running.version,
      namespace: running.namespace,
      param: {
        ...param,
        // config_commit_id: setting === 1 ? '@latest' : configCommitId,
      },
    };
    try {
      await postTasks(postdata);
      message.success('重启中');
      autoLoadApplication();
    } catch (error) {
      message.error(error.message || '重启失败');
    }
  };

  // 恢复
  const resumeRunning = async (param) => {
    const postdata = {
      clusterName,
      // 需修改
      param,
      appId: application.id,
      envName: filterInfo.envname,
      action: 'resume',
      version: running.version,
      namespace: running.namespace,
    };
    try {
      await postTasks(postdata);
      message.success('恢复中');
      autoLoadApplication();
    } catch (error) {
      message.error(error.message || '恢复失败');
    }
  };

  // 暂停
  const stopRunning = async () => {
    const postdata = {
      clusterName,
      appId: application.id,
      envName: filterInfo.envname,
      action: 'stop',
      version: running.version,
      namespace: running.namespace,
      param: {},
    };
    try {
      await postTasks(postdata);
      message.success('暂停中');
      autoLoadApplication();
    } catch (error) {
      message.error(error.message || '暂停失败');
    }
  };

  const deleteRunning = async () => {
    const postdata = {
      clusterName,
      appId: application.id,
      envName: filterInfo.envname,
      action: 'delete',
      version: running.version,
      namespace: running.namespace,
      param: {},
    };
    try {
      await postTasks(postdata);
      message.success('删除中');
      autoLoadApplication();
    } catch (error) {
      message.error(error.message || '删除失败');
    }
  };

  const fullCanaryRunning = async () => {
    const data = await getLatestTasks(
      application.id,
      filterInfo.envname,
      clusterName
    );
    if (data && data.param) {
      const postdata = {
        clusterName,
        appId: application.id,
        envName: filterInfo.envname,
        action: 'full_canary_deploy',
        version: running.version,
        namespace: running.namespace,
        param: data.param,
      };
      try {
        await postTasks(postdata);
        message.success('全量发布中');
        autoLoadApplication();
      } catch (error) {
        message.error(error.message || '全量发布失败');
      }
    }
  };

  const restartConfirm = () => {
    confirm({
      title: `确定重启${application.name}吗?`,
      icon: <ExclamationCircleOutlined />,
      // content: (
      //   <Form form={form}>
      //     <Form.Item name="setting" label="配置">
      //       <Select>
      //         <Option value={0}> 不使用 </Option>
      //         <Option value={1}> 使用最新 </Option>
      //         <Option value={2}> 自定义 </Option>
      //       </Select>
      //     </Form.Item>
      //     <Form.Item
      //       noStyle
      //       shouldUpdate={(prevValues, currentValues) =>
      //         prevValues.setting !== currentValues.setting
      //       }
      //     >
      //       {({ getFieldValue }) =>
      //         getFieldValue('setting') === 2 && (
      //           <Form.Item
      //             name="config_commit_id"
      //             label="配置ID"
      //             rules={[{ required: true, message: '请填写配置ID' }]}
      //           >
      //             <Input placeholder="请输入配置ID" />
      //           </Form.Item>
      //         )
      //       }
      //     </Form.Item>
      //     {projectInfo?.config_rename_prefixes?.length > 0 && (
      //       <Form.Item
      //         noStyle
      //         shouldUpdate={(prevValues, currentValues) =>
      //           prevValues.setting !== currentValues.setting
      //         }
      //       >
      //         {({ getFieldValue }) =>
      //           !!getFieldValue('setting') && (
      //             <Form.Item
      //               name="config_rename_prefix"
      //               label="特殊配置重命名前缀"
      //             >
      //               <Select allowClear>
      //                 {map(projectInfo.config_rename_prefixes, ({ prefix }) => (
      //                   <Option key={prefix} value={prefix}>
      //                     {prefix}
      //                   </Option>
      //                 ))}
      //               </Select>
      //             </Form.Item>
      //           )
      //         }
      //       </Form.Item>
      //     )}
      //     {projectInfo?.config_rename_prefixes?.length > 0 && (
      //       <Form.Item
      //         noStyle
      //         shouldUpdate={(prevValues, currentValues) =>
      //           prevValues.setting !== currentValues.setting
      //         }
      //       >
      //         {({ getFieldValue }) =>
      //           !!getFieldValue('setting') && (
      //             <Form.Item
      //               name="config_rename_mode"
      //               label="特殊配置重命名模式"
      //             >
      //               <Select allowClear>
      //                 {map(projectInfo.config_rename_modes, (mode) => (
      //                   <Option key={mode.enum} value={mode.enum}>
      //                     {mode.name}
      //                   </Option>
      //                 ))}
      //               </Select>
      //             </Form.Item>
      //           )
      //         }
      //       </Form.Item>
      //     )}
      //   </Form>
      // ),
      onOk() {
        return restartRunning();
      },
    });
  };

  const operateConfirm = async (type) => {
    let operateeName = '';
    let lastParam = {};
    let title = '';
    switch (type) {
      case 'delete': {
        operateeName = '删除';
        title = `确定要删除应用${application.name}的部署版本${
          runningData.version || ' '
        }吗?`;
        break;
      }
      case 'stop': {
        operateeName = '暂停';
        break;
      }
      case 'resume': {
        const result = await getLatestTasks(
          application.id,
          filterInfo.envname,
          clusterName
        );
        if (result && result.param && result.param.min_pod_count) {
          lastParam = {
            min_pod_count: result.param.min_pod_count,
          };
        }
        operateeName = '恢复';
        break;
      }
      default:
        break;
    }

    confirm({
      title: title || `确定${operateeName}${application.name}吗?`,
      icon: <ExclamationCircleOutlined />,
      onOk() {
        switch (type) {
          case 'delete':
            deleteRunning();
            break;
          case 'stop':
            stopRunning();
            break;
          case 'resume':
            // 需提前加载上一次配置
            resumeRunning(lastParam);
            break;
          default:
            break;
        }
      },
    });
  };

  const openPodsLog = async (data) => {
    try {
      const result = await getRunningPodLog({
        clusterName,
        version: running.version,
        namespace: running.namespace,
        envName: filterInfo.envname,
        podName: data.name,
        containerName: projectInfo.name + "-" + application.name
      });
      InfoModal({
        content: result || '',
        title: '日志',
      });
    } catch (error) {
      message.error(JSON.stringify(error.message) || '没找到...');
    }
  };

  const deleteJobByPodname = async (podName) => {
    confirm({
      title: `确定?`,
      icon: <ExclamationCircleOutlined />,
      onOk() {
        return deleteJobNow({
          clusterName,
          namespace: running.namespace,
          envName: filterInfo.envname,
          podName: podName,
          appId: application.id,
        });
      },
    });
  };

  const deleteJobNow = async (data) => {
    try {
      await deleteJob(data);
      autoLoadApplication();
    } catch (error) {
      message.error(JSON.stringify(error.message) || '没找到...')
    }
  }
  // 暂停发布中的镜像
  const handleStopImage = async (image, suspend) => {
    try {
      await setImageStatusStop(image.task_id, suspend);
      handleStopTaskImage();
    } catch (error) {
      message.error(error.message);
    }
  };

  // cronjob立即执行
  const actionAtOnce = async () => {
    const data = await getLatestTasks(
      application.id,
      filterInfo.envname,
      clusterName
    );
    if (data && data.param) {
      const postdata = {
        appId: application.id,
        envName: filterInfo.envname,
        action: 'manual_launch',
        version: running.version,
        namespace: running.namespace,
        clusterName,
        param: data.param,
      };
      try {
        await postTasks(postdata);
        message.success('cronjob开始执行');
        autoLoadApplication();
      } catch (error) {
        message.error(error.message || 'cronjob执行失败');
      }
    }
  };

  const confirmAtOnce = () => {
    if (runningData.is_suspend) {
      actionAtOnce();
    } else {
      confirm({
        content:
          '立即执行和当前cronjob可能存在并发执行的可能，是否确认立即执行？',
        onOk: actionAtOnce,
      });
    }
  };

  const menu = (
    <Menu>
      {running.config_url && (
        <Menu.Item>
          <a
            href={running.config_url}
            target="_blank"
            rel="noopener noreferrer"
          >
            配置
          </a>
        </Menu.Item>
      )}
      {includes([SERVICE, WORKER], application.type) &&
        running.task_status === 'success' &&
        hasMaintainerPermission(memberRole, filterInfo.envname) && (
          <Menu.Item>
            <a onClick={openUpdateHpaModal}>配置弹性伸缩</a>
          </Menu.Item>
        )}
      <Menu.Item>
        <a onClick={openRunningInfoLog}>发布详情日志</a>
      </Menu.Item>
      {application.type === CRON_JOB && (
        <>
          <Menu.Item>
            <a onClick={openCronJobInfo}>详情</a>
          </Menu.Item>

          {hasMaintainerPermission(memberRole, filterInfo.envname) && (
            <Menu.Item>
              <a onClick={confirmAtOnce}>立即执行</a>
            </Menu.Item>
          )}
        </>
      )}
      {running.pod_monitor_url && application.type !== ONE_TIME_JOB && (
        <Menu.Item>
          <a
            target="_blank"
            rel="noopener noreferrer"
            href={running.pod_monitor_url}
          >
            监控
          </a>
        </Menu.Item>
      )}
      {runningData.pod_monitor_url && application.type === ONE_TIME_JOB && (
        <Menu.Item>
          <a
            target="_blank"
            rel="noopener noreferrer"
            href={runningData.pod_monitor_url}
          >
            监控
          </a>
        </Menu.Item>
      )}

      {!includes([CRON_JOB, ONE_TIME_JOB], application.type) &&
        hasDeveloperPermission(memberRole, filterInfo.envname) && running.task_display_icon !== 'resume'  && (
          <Menu.Item>
            <a onClick={restartConfirm}>重启</a>
          </Menu.Item>
        )}

      {application.type !== ONE_TIME_JOB &&
        hasMaintainerPermission(memberRole, filterInfo.envname) &&
        (running.task_display_icon === 'resume' ? (
          <Menu.Item>
            <a onClick={() => operateConfirm('resume')}>恢复</a>
          </Menu.Item>
        ) : (
          <Menu.Item>
            <a onClick={() => operateConfirm('stop')}>暂停</a>
          </Menu.Item>
        ))}

      {hasMaintainerPermission(memberRole, filterInfo.envname) && (
        <Menu.Item>
          <a onClick={() => operateConfirm('delete')}>删除</a>
        </Menu.Item>
      )}
    </Menu>
  );

  const PodsRender = ({ pods }) => {
    return (
      pods &&
      pods.map((item) => (
        <Row
          key={item.name}
          style={{ marginBottom: '5px', marginRight: '10px' }}
        >
          <Col span={15}>
            <Divider type="vertical" /> {item.name}{' '}
          </Col>
          <Col span={4}>
            {item.phase ? <JobsPodStatus status={item.phase} /> : null}

            <Divider type="vertical" style={{ margin: '0 2px' }} />
            {application.type === ONE_TIME_JOB ? item.create_time : item.age+' Ago'}
            <Divider type="vertical" style={{ margin: '0 2px' }} />
            <Badge
              count={item.restart_count}
              style={{
                margin: '0px 2px',
                height: '13px',
                padding: '0 2px',
                fontSize: '10px',
                lineHeight: '13px',
                minWidth: '13px',
              }}
            />
          </Col>
          <Col span={5} className={styles.textRight}>
            <Space align="center">
              <a onClick={() => openPodsLog(item)}> 日志 </a>
              <a
                onClick={() => {
                  handleChangeVisible(true);
                  handleChangeActPod(item);
                  handleChangeActRunning(running);
                }}
              >
                状态
              </a>

              {/*<a*/}
              {/*  target="_blank"*/}
              {/*  rel="noopener noreferrer"*/}
              {/*  href={item.shell_url}*/}
              {/*>*/}
              {/*  {' '}*/}
              {/*  shell{' '}*/}
              {/*</a>*/}
              {item.phase !== 'Failed' && projectInfo?.language === 'Go' && (
                <a
                  onClick={() => handleCreatePprof(item, running, application)}
                >
                  pprof
                </a>
              )}
              {(application.type === CRON_JOB || application.type === ONE_TIME_JOB) && (
                <a onClick={() => deleteJobByPodname(item.name)}> 删除 </a>
              )}
            </Space>
          </Col>
        </Row>
      ))
    );
  };

  return (
    <>
      <Row style={{ marginBottom: '5px' }}>
        <Col span={9}>
          <Badge status={running.network_traffic ? 'success' : 'default'} />
          {`${running.namespace ? `${running.namespace}.` : ''}${
            running.version
          }`}
        </Col>
        <Col span={5}> {running.create_time} </Col>
        {/* 13 11 */}
        <Col span={6}>
          <Tooltip
            title={`镜像版本：${running.image_version}`}
            trigger={['click', 'hover']}
          >
            <QuestionCircleOutlined />
          </Tooltip>
          &nbsp;&nbsp;
          <RunningStatus
            status={running.task_status}
            task_display_icon={running.task_display_icon}
          />
          <Divider type="vertical" style={{ margin: '0 2px' }} />
          {running.task_display_icon === 'canary_continue' ? (
            <Button size="small" type="primary" onClick={fullCanaryRunning}>
              {' '}
              继续{' '}
            </Button>
          ) : (
            running.task_status_display
          )}
          {application.type !== ONE_TIME_JOB ? (
            <>
              {['underway'].includes(running.task_display_icon) && (
                <>
                  &nbsp;
                  <Button
                    type="primary"
                    size="small"
                    danger
                    onClick={() => handleStopImage(running, true)}
                  >
                    暂停
                  </Button>
                </>
              )}
              {running.task_suspend === true && (
                <>
                  &nbsp;
                  <Button
                    type="primary"
                    size="small"
                    onClick={() => handleStopImage(running, false)}
                  >
                    恢复
                  </Button>
                </>
              )}
            </>
          ) : null}
        </Col>

        <Col span={4} className={styles.textRight}>
          <Space align="center">
            <Button size="small" type="primary" className={styles.podButton}>
              {' '}
              实例{' '}
            </Button>
            <Button
              size="small"
              type="primary"
              className={styles.podButton}
              onClick={() => setTaskDetailVisible(true)}
            >
              {' '}
              参数{' '}
            </Button>
            <Dropdown overlay={menu}>
              <a
                className="ant-dropdown-link"
                onClick={(e) => e.preventDefault()}
              >
                更多
                <DownOutlined />
              </a>
            </Dropdown>
          </Space>
        </Col>
      </Row>
      {taskDetailVisible && (
        <TaskParamsDetail
          onHide={() => setTaskDetailVisible(false)}
          application={application}
          filterInfo={filterInfo}
          running={running}
          clusterName={clusterName}
          projectInfo={projectInfo}
        />
      )}
      <UpdateHpaModal
        visible={updateHpaVisible}
        application={application}
        filterInfo={filterInfo}
        onSubmit={onUpdateHpaSubmit}
        onCancel={() => setUpdateHpaVisible(false)}
        clusterName={clusterName}
      />
      <Spin spinning={loading} size="small">
        {runngingPods &&
          runngingPods.map((item) => {
            return (
              <Row key={item.name} style={{ marginBottom: '5px' }}>
                <Col span={15}> {item.name} </Col>
                <Col span={4}>
                  {item.phase ? <PodsStatus status={item.phase} /> : null}
                  <Divider type="vertical" style={{ margin: '0 2px' }} />
                  {item.age}
                  <Divider type="vertical" style={{ margin: '0 2px' }} />
                  <Badge
                    count={item.restart_count}
                    style={{
                      margin: '0px 2px',
                      height: '13px',
                      padding: '0 2px',
                      fontSize: '10px',
                      lineHeight: '13px',
                      minWidth: '13px',
                    }}
                  />
                </Col>
                <Col span={5} className={styles.textRight}>
                  <Space align="center">
                    <a onClick={() => openPodsLog(item)}> 日志 </a>
                    <a
                      onClick={() => {
                        handleChangeVisible(true);
                        handleChangeActPod(item);
                        handleChangeActRunning(running);
                      }}
                    >
                      状态
                    </a>

                    {/*<a*/}
                    {/*  target="_blank"*/}
                    {/*  rel="noopener noreferrer"*/}
                    {/*  href={item.shell_url}*/}
                    {/*>*/}
                    {/*  shell*/}
                    {/*</a>*/}
                    {projectInfo?.language === 'Go' && (
                      <a
                        onClick={() =>
                          handleCreatePprof(item, running, application)
                        }
                      >
                        pprof
                      </a>
                    )}
                  </Space>
                </Col>
              </Row>
            );
          })}
        {runngingJobs &&
          runngingJobs.map((item) => {
            return (
              <div key={item.name}>
                {application.type === ONE_TIME_JOB ? null : (
                  <Row>
                    <Col span={13}> {item.name} </Col>
                    <Col span={1}>
                      {item.succeeded_count}/{item.need_complete_count}
                    </Col>

                    <Col span={6}>
                      {item.start_time}
                      {application.type === CRON_JOB ? (
                        <>
                          {item.failed_count > 0 ? (
                            <>
                              <CloseCircleFilled style={{ color: 'red' }} />
                              失败
                            </>
                          ) : null}{' '}
                          {item.completion_time ? (
                            <div>{item.completion_time}</div>
                          ) : null}
                        </>
                      ) : null}
                    </Col>
                  </Row>
                )}
                <div>
                  <PodsRender pods={item.pods} />
                </div>
              </div>
            );
          })}
      </Spin>
    </>
  );
};

export default RunItem;

import { Card, Skeleton, message, Pagination } from 'antd';
import { connect, useLocation } from 'umi';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import React, { useState, useEffect, useRef, useMemo } from 'react';
import ProjectInfo from './components/ProjectInfo';
import NewImages from './components/NewImages';
import AppFilter from './components/AppFilter';
import CommonParams from './components/CommonParams';
import AppContainer from './components/AppContainer';
import CreatePodPprof from './components/CreatePodPprof';
import CreateApplication from './components/CreateApplication';
import BatchActionTasks from './components/BatchActionTasks';
import {
  buildImage,
  createApp,
  getProjectimageTag,
  postBatchAction,
  getNodeLabels,
  collectProject,
  uncollectProject,
} from './service';
import { applicationActions } from '../../../utils/commonData';
import './style.less';

const Application = (props) => {
  const timer = useRef();
  const location = useLocation();
  const {
    dispatch,
    projectInfo,
    applications,
    allApplications,
    total,
    allLabels,
    loadingInit,
    imageArgTemplates,
    memberRole,
    match: {
      params: { projectId },
    },
    location: {
      query: { envname, keyword },
    },
  } = props;
  const [createApplicationModal, setCreateApplicationModal] = useState(false);
  const [batchActionTaskVisible, setBatchActionTaskVisible] = useState(false);
  const [filterInfo, setFilterInfo] = useState({
    envname: envname || 'test',
    keyword: keyword || '',
  });
  const [pprofVisible, setPprofVisible] = useState(false); // pprof弹框
  const [actApplication, setActApplicaiton] = useState({}); // 点前点击的application
  const [actRunning, setActRunning] = useState({}); // 当前点击的running
  const [actPod, setActPod] = useState({}); // 当前点击的pod
  const [newImageTags, setNewImageTags] = useState([]); // 镜像列表
  const [batchActionLoading, setBatchActionLoading] = useState(false);
  const [nodeLabels, setNodeLabels] = useState([]); // 高级选项下拉可选内容
  const [collect, setCollect] = useState(false); // 是否被收藏

  // 分页
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(5);

  // 刷新内容
  const [reloadingApplications, setReloadingApplications] = useState(false);

  // 获取当前项目镜像列表
  const getProjectimageTagDate = () => {
    setBatchActionLoading(true);
    getProjectimageTag(projectId).then((res) => {
      setNewImageTags(res.list || []);
      setBatchActionLoading(false);
    });
  };

  const filters = useMemo(() => {
    const { keyword: kw, type } = location.query ?? {};
    return {
      keyword: kw,
      type,
    };
  }, [location.query?.keyword, location.query?.type]);

  useEffect(() => {
    dispatch({
      type: 'application/initAll',
      payload: {
        page,
        project_id: projectId,
        env_name: filterInfo.envname,
      },
    });

    dispatch({
      type: 'application/init',
      payload: {
        page,
        limit: pageSize,
        project_id: projectId,
        env_name: filterInfo.envname,
        ...filters,
      },
    });
    setReloadingApplications(false);

    (async () => {
      try {
        const res = await getNodeLabels();
        setNodeLabels(res || []);
      } catch (error) {
        setNodeLabels([]);
        message.error(error.message);
      }
    })();

    return () => {
      dispatch({ type: 'application/clear' });
    };
  }, []);

  // 手动刷新
  useEffect(() => {
    if (reloadingApplications) {
      dispatch({
        type: 'application/getProjectApplication',
        payload: {
          page,
          limit: pageSize,
          project_id: projectId,
          env_name: filterInfo.envname,
          ...filters,
        },
      }).then(() => {
        setReloadingApplications(false);
      });
    }
  }, [reloadingApplications]);

  useEffect(() => {
    setCollect(projectInfo?.is_fav || false);
  }, [projectInfo]);

  const mounted = useRef();
  useEffect(() => {
    if (!mounted.current) {
      mounted.current = true;
    } else {
      setReloadingApplications(true);
    }
  }, [page, pageSize]);

  const m = useRef();
  useEffect(() => {
    if (!m.current) {
      m.current = true;
    } else if (page > 1) setPage(1);
    else setReloadingApplications(true);
  }, [filters]);

  // 收藏/取消收藏项目
  const handleColllectProject = () => {
    if (!collect) {
      try {
        collectProject(projectId);
        message.success('已收藏');
        setCollect(true);
      } catch (error) {
        message.error(error.message || '收藏失败');
      }
    } else {
      try {
        uncollectProject(projectId);
        message.success('已取消收藏');
        setCollect(false);
      } catch (error) {
        message.error(error.message || '取消收藏失败');
      }
    }
  };

  const doCreateApplication = async (values) => {
    // 同一项目一次性任务数量不超过3个
    const oneTimeJobCount = (applications || []).filter(
      (i) => i.type === 'OneTimeJob'
    );
    if (oneTimeJobCount >= 3 && values.type === 'OneTimeJob') {
      message.destroy();
      message.error('同一项目一次性任务数量不超过3个');
      return;
    }

    try {
      await createApp({ ...values, project_id: projectId });
      setReloadingApplications(true);
      setCreateApplicationModal(false);
    } catch (error) {
      message.error(error.message || '创建失败');
      setCreateApplicationModal(false);
    }
  };

  const handleCreatePprof = (pod, running, application) => {
    setActPod(pod);
    setActRunning(running);
    setActApplicaiton(application);
    setPprofVisible(true);
  };
  // 批量操作应用
  const batchAction = () => {
    setBatchActionTaskVisible(true);
    getProjectimageTagDate();
  };

  // 确认批量操作
  const handleSubmitBatchAction = async (param) => {
    const data = {
      project_id: projectId,
      env_name: filterInfo.envname,
      ...param,
    };

    try {
      await postBatchAction(data);
      message.success(
        `批量操作成功,正在进行批量${
          applicationActions.find((i) => i.value === param.action)?.name
        }操作`
      );
      setBatchActionTaskVisible(false);
      // 刷新项目数据
      if (timer.current) clearTimeout(timer.current);
      timer.current = setTimeout(() => {
        setReloadingApplications(true);
      }, 1500);
    } catch (error) {
      message.error(error.message || '操作失败');
    }
  };

  useEffect(() => {
    dispatch({
      type: 'application/getProjectClusters',
      payload: {
        projectId,
      },
    });
  }, [projectId]);

  return (
    <PageHeaderWrapper
      title={false}
      content={
        <ProjectInfo
          projectId={projectId}
          filterInfo={filterInfo}
          data={projectInfo}
          allLabels={allLabels}
          collect={collect}
          memberRole={memberRole}
          handleColllectProject={() => handleColllectProject()}
        />
      }
      extraContent={
        <NewImages
          projectId={projectId}
          filterInfo={filterInfo}
          buildImage={buildImage}
          memberRole={memberRole}
          imageArgTemplates={imageArgTemplates}
          imageArgs={
            projectInfo && projectInfo.image_args
              ? projectInfo.image_args
              : null
          }
        />
      }
    >
      {/* <CommonParams
        projectId={projectId}
        memberRole={memberRole}
        filterInfo={filterInfo}
      /> */}
      <Card
        bordered={false}
        loading={loadingInit} // loading
      >
        <AppFilter
          projectId={projectId}
          createApp={setCreateApplicationModal}
          filterInfo={filterInfo}
          setFilterInfo={setFilterInfo}
          setReloadingApplications={setReloadingApplications}
          batchAction={batchAction}
          memberRole={memberRole}
        />

        {reloadingApplications ? (
          <Skeleton paragraph={{ rows: 10 }} />
        ) : (
          applications.length > 0 && (
            <div
              style={{
                overflow: 'auto',
              }}
            >
              {applications.map((item) => {
                return (
                  <AppContainer
                    key={`app_${item.id}`}
                    projectInfo={projectInfo}
                    nodeLabels={nodeLabels}
                    filterInfo={filterInfo}
                    setReloadingApplications={setReloadingApplications}
                    handleCreatePprof={(pod, running, application) =>
                      handleCreatePprof(pod, running, application)
                    }
                    projectId={projectId}
                    application={item}
                    memberRole={memberRole}
                  />
                );
              })}
            </div>
          )
        )}
        <Pagination
          size="small"
          style={{ textAlign: 'right' }}
          total={total}
          current={page}
          pageSize={pageSize}
          showQuickJumper
          showSizeChanger
          showTotal={(t, range) => `第${range[0]}-${range[1]}条 / 总共 ${t} 条`}
          pageSizeOptions={['5', '10', '15', '20']}
          onChange={(p) => setPage(p)}
          onShowSizeChange={(current, size) => {
            setPage(1);
            setPageSize(size);
          }}
        />
      </Card>

      {/* 创建app弹出层 */}
      <CreateApplication
        onSubmit={doCreateApplication}
        onCancel={() => {
          setCreateApplicationModal(false);
        }}
        projectInfo={projectInfo}
        modalVisible={createApplicationModal}
      />

      {/* pprof弹出层 */}
      <CreatePodPprof
        visible={pprofVisible}
        actPod={actPod}
        actRunning={actRunning}
        actApplication={actApplication}
        projectInfo={projectInfo}
        filterInfo={filterInfo}
        onCancel={(bol) => setPprofVisible(bol)}
        onSubmit={() => setPprofVisible(false)}
      />

      {/* 批量操作 */}
      <BatchActionTasks
        visible={batchActionTaskVisible}
        applications={allApplications}
        filterInfo={filterInfo}
        loading={batchActionLoading}
        newImageTags={newImageTags}
        projectId={projectId}
        projectInfo={projectInfo}
        onSubmit={(query) => handleSubmitBatchAction(query)}
        onCancel={() => setBatchActionTaskVisible(false)}
      />
    </PageHeaderWrapper>
  );
};

export default connect(
  ({
    application: {
      projectInfo,
      applications,
      allApplications,
      total,
      allLabels,
      imageArgTemplates,
      memberRole,
    },
    loading,
  }) => ({
    total,
    projectInfo,
    // images,
    applications,
    allApplications,
    allLabels,
    imageArgTemplates,
    memberRole,
    // imageTags,
    loadingInit: loading.effects['application/init'],
    loadingProjectInfo: loading.effects['application/getProjectInfo'],
    // loadingProjectImagesInfo: loading.effects['application/getProjectImagesInfo'],
    getProjectApplication: loading.effects['application/getProjectApplication'],
  })
)(Application);

import {
  Button,
  Space,
  Divider,
  Row,
  Col,
  message,
  Skeleton,
  Tooltip,
} from 'antd';
import { throttle } from 'lodash';
import { ReloadOutlined } from '@ant-design/icons';
import React, { useState, useEffect, useCallback, useMemo } from 'react';
import showBuildLogModal from '@/utils/showBuildLogModal';
import StatusIcon from '@/components/StatusIcon';
import WordTip from '@/components/Common/WordTip';
import { useEventCallback } from '@/lib/useEventCallback';
import styles from './style.less';
import BuildImage from './BuildImage';
import { getProjectImagesInfo, deleteImagesBuild } from '../service';
import { hasDeveloperPermission } from '../utils/hasPermission';

const NewImages = ({
  projectId,
  buildImage,
  memberRole,
  imageArgs,
  imageArgTemplates,
  filterInfo,
}) => {
  const [createModalVisible, handleModalVisible] = useState(false);
  const [reloadingImage, setReloadingImage] = useState(false);

  const [imagesData, setImagesData] = useState([]);

  const getImagesData = useEventCallback(async () => {
    const res = await getProjectImagesInfo(projectId);
    setImagesData(res.list);
  });

  const loading = async () => {
    setReloadingImage(false);
    await getImagesData();
    setReloadingImage(true);
  };

  useEffect(() => {
    loading();
  }, []);

  const throttled = useMemo(
    () => throttle(getImagesData, 3000, { leading: false }),
    []
  );

  useEffect(() => {
    const isLoading = imagesData.some(
      (item) =>
        item.status === 'BUILDING' ||
        item.status === 'LOADING' ||
        item.status === ''
    );
    if (isLoading) {
      throttled();
    }
  }, [imagesData]);

  useEffect(() => {
    return () => {
      throttled.cancel();
    };
  }, []);

  const doBuild = async (values) => {
    try {
      await buildImage(values);
      message.success('镜像构建任务,创建成功,请耐心等待后台构建...');
      handleModalVisible(false);
      // 触发重新获取
      getImagesData();
    } catch (e) {
      message.error(e.message || '镜像构建任务，创建失败');
    }
  };

  const openImageLog = useCallback(
    (buildId) => showBuildLogModal(projectId, buildId),
    [projectId]
  );

  const stopBuild = async (buildId) => {
    try {
      await deleteImagesBuild(projectId, buildId);
      message.success('镜像构建已停止,你可尝试新的镜像构建');
      handleModalVisible(false);
      // 触发重新获取
      getImagesData();
    } catch (e) {
      message.error(e.message || '镜像构建停止失败');
    }
  };

  return (
    <div className={styles.newDocker}>
      <div style={{ marginTop: '10px' }}>
        <Row style={{ marginBottom: '5px' }} className={styles.header}>
          <Col span={3} style={{ padding: '5px' }}>
            <Space align="center">
              <span>
                最新镜像
                <Divider type="horizontal" className={styles.titleDivider} />
              </span>
            </Space>
          </Col>
          <Col span={17}>
            <Space align="center">
              <Button
                onClick={loading}
                className={styles.loadBution}
                icon={<ReloadOutlined />}
              >
                刷新
              </Button>
            </Space>
          </Col>
          <Col span={4}>
            <Space align="center">
              <a href={`/project/${projectId}/image`}>{`镜像列表 >>`} </a>
            </Space>
          </Col>
        </Row>
        {reloadingImage ? (
          <>
            {imagesData &&
              imagesData.map((item, index) => {
                return (
                  <Row key={index} style={{ marginBottom: '5px' }}>
                    <Col span={6} style={{ overflowWrap: 'break-word' }}>
                      <Tooltip placement="topLeft" title={item.branch_name}>
                        {item.branch_name}
                      </Tooltip>
                    </Col>
                    <Col span={4} style={{ paddingLeft: '10px' }}>
                      <Tooltip placement="topLeft" title={item.last_comment}>
                        {item.commit_id}
                      </Tooltip>
                    </Col>
                    <Col span={6}>{item.create_time} </Col>
                    <Col span={2}>
                      <WordTip title={item.user_profile?.name ?? ''} />
                    </Col>
                    <Col span={3}>
                      <a onClick={() => openImageLog(item.build_id)}>
                        构建日志
                      </a>
                    </Col>
                    <Col span={3}>
                      <StatusIcon status={item.status} />
                      {item.status === 'BUILDING' ||
                      item.status === 'LOADING' ||
                      item.status === '' ? (
                        <>
                          <Divider
                            type="vertical"
                            style={{ margin: '0 2px' }}
                          />
                          <Button
                            size="small"
                            type="primary"
                            onClick={() => stopBuild(item.build_id)}
                          >
                            停止
                          </Button>
                        </>
                      ) : null}
                    </Col>
                  </Row>
                );
              })}
          </>
        ) : (
          <Skeleton paragraph={{ rows: 1 }} />
        )}
      </div>
      {hasDeveloperPermission(memberRole, filterInfo.envname) && (
        <Button type="primary" onClick={() => handleModalVisible(true)}>
          构建镜像
        </Button>
      )}
      <BuildImage
        onSubmit={doBuild}
        onCancel={() => {
          handleModalVisible(false);
        }}
        modalVisible={createModalVisible}
        projectId={projectId}
        imageArgs={imageArgs}
        imageArgTemplates={imageArgTemplates}
      />
    </div>
  );
};

export default NewImages;

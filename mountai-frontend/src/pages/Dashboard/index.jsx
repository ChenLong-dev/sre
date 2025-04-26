import { Avatar, Card, Col, List, Row, Pagination, message } from 'antd';
import { StarFilled } from '@ant-design/icons';
import React, { Component } from 'react';
import { Link, connect } from 'umi';
// import { PageHeaderWrapper } from '@ant-design/pro-layout';
import moment from 'moment';
import { getCollectData, uncollectProject } from './service';

import styles from './style.less';

class Dashboard extends Component {
  constructor(props) {
    super(props);
    this.state = {
      collectLoading: false,
      collectData: [],
      collectTotal: 0,
      collectLength: 0,
      collectPage: 1,
      collectLimit: 5,
    };
  }

  componentDidMount() {
    const { dispatch } = this.props;
    dispatch({
      type: 'dashboard/init',
    });
    this.getCollects();
  }

  componentWillUnmount() {
    const { dispatch } = this.props;

    dispatch({
      type: 'dashboard/clear',
    });
  }

  getCollects = async () => {
    const { collectPage, collectLimit } = this.state;
    const query = { page: collectPage, limit: collectLimit };
    this.setState({ collectLoading: true });
    try {
      const res = await getCollectData(query);
      this.setState({
        collectData: res?.list || [],
        collectLength: (res?.list || []).length,
        collectTotal: res.count || 0,
      });
    } catch (error) {
      message.error(error.message || '获取收藏列表失败');
    } finally {
      this.setState({ collectLoading: false });
    }
  };

  // 取消收藏
  handleCollectProject = async (item) => {
    const { collectPage, collectLength } = this.state;
    try {
      await uncollectProject(item.project.id);
      message.success('已取消收藏');
      if (collectPage > 1 && collectLength === 1) {
        this.setState({ collectPage: collectPage - 1 }, () =>
          this.getCollects()
        );
      } else {
        this.getCollects();
      }
    } catch (error) {
      message.error(error.message || '取消收藏失败');
    }
  };

  renderCollectList = (item) => {
    return (
      <List.Item
        actions={[
          <StarFilled
            onClick={() => this.handleCollectProject(item)}
            style={{ fontSize: '18px', cursor: 'pointer', color: '#fadb14' }}
          />,
        ]}
      >
        <List.Item.Meta
          avatar={<Avatar src={item.user.avatar_url} />}
          title={
            <a
              className={styles.event}
              href={`/project/${item.project.id}/application`}
            >
              {item.project.name}
            </a>
          }
          description={`${item.project.desc}`}
        />
      </List.Item>
    );
  };

  renderActivities = (item) => {
    const events = item.template.split(/@\{([^{}]*)\}/gi).map((key) => {
      if (item[key]) {
        return (
          <a href={item[key].link} key={item[key].name}>
            {item[key].name}
          </a>
        );
      }

      return key;
    });
    return (
      <List.Item key={item.id}>
        <List.Item.Meta
          avatar={<Avatar src={item.user.avatar} />}
          title={
            <span>
              <a className={styles.username}>{item.user.name}</a>
              &nbsp;
              <span className={styles.event}>{events}</span>
            </span>
          }
          description={
            <span className={styles.datetime} title={item.updatedAt}>
              {moment(item.updatedAt).fromNow()}
            </span>
          }
        />
      </List.Item>
    );
  };

  renderActivitiesLog = (item) => {
    return (
      <List.Item key={item.id}>
        <List.Item.Meta
          avatar={<Avatar src={item.operator_avatar_url} />}
          title={
            <span>
              <a
                className={styles.event}
                href={`/project/${item.project_id}/application`}
              >
                {item.project_name}项目
              </a>
              的&nbsp;
              <span className={styles.event}>
                <a
                  href={`/project/${item.project_id}/application?keyword=${item.app_name}`}
                >
                  {item.app_name}应用
                </a>
                在&nbsp;
                <a
                  href={`/project/${item.project_id}/application?envname=${item.env_name}`}
                >
                  {item.env_name}环境
                </a>
                &nbsp; 被<a> {item.operator_name} </a>
                执行了
                <a> {item.action_display} </a>
              </span>
            </span>
          }
          description={
            <span className={styles.datetime} title={item.create_time}>
              {moment(item.create_time).fromNow()}
            </span>
          }
        />
      </List.Item>
    );
  };

  render() {
    const {
      // activities,
      // activitiesLoading,
      activitiesLogLoading,
      teamLoading,
      teams,
      usedProjects,
      activitiesLogs,
    } = this.props;

    const {
      collectData,
      collectLoading,
      collectTotal,
      collectPage,
      collectLimit,
    } = this.state;

    // if (!currentUser || !currentUser.userid) {
    //   return null;
    // }

    return (
      // <PageHeaderWrapper
      //   // content={<PageHeaderContent currentUser={currentUser} />}
      //   // extraContent={<ExtraContent />}
      // >
      <Row gutter={24}>
        <Col xl={16} lg={24} md={24} sm={24} xs={24}>
          <Card
            className={styles.projectList}
            style={{
              marginBottom: 24,
            }}
            title="最近使用"
            bordered={false}
            extra={<Link to="/projects/list">全部项目</Link>}
            // loading={initLoading}
            bodyStyle={{
              padding: 0,
            }}
          >
            {usedProjects.map((item) => (
              <Card.Grid className={styles.projectGrid} key={item.id}>
                <Card
                  bodyStyle={{
                    padding: 0,
                  }}
                  bordered={false}
                >
                  <Card.Meta
                    title={
                      <div className={styles.cardTitle}>
                        {/* <Avatar size="small" src={item.logo} /> */}
                        <Link to={`/project/${item.id}/application`}>
                          {item.name}
                        </Link>
                      </div>
                    }
                    description={item.desc || ' '}
                  />
                  <div className={styles.projectItemContent}>
                    <Link to="/team/list">{item.team_name || ''}</Link>
                    {item.task_create_time && (
                      <span
                        className={styles.datetime}
                        title={item.task_create_time}
                      >
                        {item.task_create_time}
                        {/* task_create_time
                          {moment(item.updatedAt).fromNow()} */}
                      </span>
                    )}
                  </div>
                </Card>
              </Card.Grid>
            ))}
          </Card>
          <Card
            bodyStyle={{
              padding: 0,
            }}
            bordered={false}
            className={styles.activeCard}
            title="个人动态"
            extra={<Link to="/projects/dynamic">更多</Link>}
            loading={activitiesLogLoading}
          >
            {/* <List
                loading={activitiesLogLoading}
                renderItem={item => this.renderActivities(item)}
                dataSource={activities}
                className={styles.activitiesList}
                size="large"
              /> */}

            <List
              loading={activitiesLogLoading}
              renderItem={(item) => this.renderActivitiesLog(item)}
              dataSource={activitiesLogs}
              className={styles.activitiesList}
              size="large"
            />
          </Card>
        </Col>
        <Col xl={8} lg={24} md={24} sm={24} xs={24}>
          <Card
            title="我的收藏"
            style={{
              marginBottom: 24,
            }}
            bodyStyle={{
              padding: 0,
            }}
            bordered={false}
          >
            <List
              dataSource={collectData}
              loading={collectLoading}
              className={styles.activitiesList}
              renderItem={(item) => this.renderCollectList(item)}
            />
            {collectData.length > 0 ? (
              <Pagination
                size="small"
                style={{
                  textAlign: 'right',
                  marginBottom: '20px',
                  marginRight: '20px',
                }}
                total={collectTotal}
                current={collectPage}
                pageSize={collectLimit}
                showQuickJumper
                showSizeChanger
                showTotal={(total, range) =>
                  `第${range[0]}-${range[1]}条 / 总共 ${total} 条`
                }
                pageSizeOptions={['5', '10', '15', '20']}
                onChange={(page) =>
                  this.setState({ collectPage: page }, () => this.getCollects())
                }
                onShowSizeChange={(current, size) => {
                  this.setState({ collectPage: 1, collectLimit: size }, () =>
                    this.getCollects()
                  );
                }}
              />
            ) : null}
          </Card>
          <Card
            bodyStyle={{
              paddingTop: 12,
              paddingBottom: 12,
            }}
            bordered={false}
            title="团队项目"
            loading={teamLoading}
          >
            {/* 或写死团队list */}
            <div className={styles.members}>
              <Row gutter={48}>
                {teams.map((item) => (
                  <Col span={12} key={`members-item-${item.id}`}>
                    <Link to={`/projects/list?team_id=${item.id}`}>
                      {/* <Avatar src={item.logo} size="small" /> */}
                      <span className={styles.member}>{item.name}</span>
                    </Link>
                  </Col>
                ))}
              </Row>
            </div>
          </Card>
        </Col>
      </Row>
      // </PageHeaderWrapper>
    );
  }
}

export default connect(
  ({ dashboard: { usedProjects, teams, activitiesLogs }, loading }) => ({
    // activities,
    teams,
    activitiesLogs,
    usedProjects,
    initLoading: loading.effects['dashboard/init'],
    // activitiesLoading: loading.effects['dashboard/fetchActivitiesList'], //假数据
    teamLoading: loading.effects['dashboard/fetchTeam'],
    usedProjectsLoading: loading.effects['dashboard/fetchUsedProjects'],
    activitiesLogLoading: loading.effects['dashboard/fetchActivitiesLog'],
  })
)(Dashboard);

import React from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Card, Alert, Typography } from 'antd';
import styles from './Home.less';

const CodePreview = ({ children }) => (
  <pre className={styles.pre}>
    <code>
      <Typography.Text copyable>{children}</Typography.Text>
    </code>
  </pre>
);

export default () => (
  <PageContainer>
    <Card>
      <Alert
        message="更快更强的重型组件，已经发布。"
        type="success"
        showIcon
        banner
        style={{
          margin: -12,
          marginBottom: 24,
        }}
      />
      {/* <Typography.Text strong>
        高级表格{' '}
        <a href="https://protable.ant.design/" rel="noopener noreferrer" target="__blank">
          欢迎使用
        </a>
      </Typography.Text> */}
      <CodePreview>首页欢迎你</CodePreview>
      {/* <Typography.Text
        strong
        style={{
          marginBottom: 12,
        }}
      >
        高级布局{' '}
        <a href="https://prolayout.ant.design/" rel="noopener noreferrer" target="__blank">
          欢迎使用
        </a>
      </Typography.Text> */}
      <CodePreview>首页</CodePreview>
    </Card>
  </PageContainer>
);

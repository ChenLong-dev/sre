import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { Spin, Alert } from 'antd';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { docco } from 'react-syntax-highlighter/dist/esm/styles/hljs';
import { getImageLog } from '@/services/image';
import Styles from './index.less';

function BuildLog({ projectId, buildId }) {
  const [content, setContent] = useState();
  const [error, setError] = useState();
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let t;
    async function getLog() {
      try {
        const result = await getImageLog(projectId, buildId);
        if (result?.console_output) {
          setContent(result?.console_output);
        }
        if (!result.status) {
          t = setTimeout(() => {
            getLog();
          }, 15000);
        }
      } catch (e) {
        setError(e.message);
      }
      setIsLoading(false);
    }
    getLog();
    return () => {
      clearTimeout(t);
    };
  }, []);

  if (isLoading) {
    return <Spin style={{ width: '100%' }} size="large" />;
  }

  if (error) {
    return <Alert message="Error" description={error} type="error" showIcon />;
  }

  return (
    <SyntaxHighlighter
      wrapLongLines
      className={Styles.logContent}
      language="shell"
      style={docco}
    >
      {content}
    </SyntaxHighlighter>
  );
}

BuildLog.propTypes = {
  projectId: PropTypes.string.isRequired,
  buildId: PropTypes.string.isRequired,
};

export default React.memo(BuildLog);

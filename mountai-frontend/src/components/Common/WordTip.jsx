import React from 'react';
import PropTypes from 'prop-types';
import { Tooltip } from 'antd';
import { QuestionCircleOutlined } from '@ant-design/icons';

const toolTipStyle = {
  maxWidth: 540,
  whiteSpace: 'pre',
};

const WordTip = ({ title, children, className }) => {
  return (
    <Tooltip title={title} overlayStyle={toolTipStyle}>
      {children || <QuestionCircleOutlined className={className} />}
    </Tooltip>
  );
};

WordTip.propTypes = {
  title: PropTypes.string.isRequired,
  children: PropTypes.node,
  className: PropTypes.string,
};

WordTip.defaultProps = {
  children: null,
  className: '',
};

export default React.memo(WordTip);

import { CheckCircleFilled, WarningFilled, LoadingOutlined} from '@ant-design/icons';
import { Spin } from 'antd';

const StatusIcon = ({status}) => {

  if(status === 'SUCCESS') {
    return (
      <CheckCircleFilled style={{color:'rgb(6, 170, 36)'}}/> 
    );
  }else if(status === 'BUILDING' || status === 'LOADING' ||status === '') {
    return (
      <LoadingOutlined  /> 
    );
  } else {
    return (
      <WarningFilled style={{color:'rgb(249, 10, 10)'}}/> 
    );
  }  
};

export default StatusIcon;

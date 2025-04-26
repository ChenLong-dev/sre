import { Modal } from 'antd';
import React from 'react';
import styles from './index.less';

const InfoModal = ({content, title}) => {
    // console.log('text-content', content);
    return Modal.info({
        title,
        width:1000,
        maskClosable:true,
        content: (
            <div className={styles.preLine} >
                {content}
            </div>
        ),
        onOk() {},
    });
}

export default InfoModal;

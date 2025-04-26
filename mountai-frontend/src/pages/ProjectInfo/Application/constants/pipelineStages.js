import { map } from 'lodash';

const PIPELINE_STAGES = {
  CODE_SCAN: 'CodeScan',
  UNIT_TEST: 'UnitTest',
  BUILD: 'Build',
  DEPLOY_FAT: 'DeployFat',
  API_TEST: 'APITest',
  DEPLOY_STG: 'DeployStg',
};

const PIPELINE_STAGE_NAMES = {
  [PIPELINE_STAGES.CODE_SCAN]: '代码扫码',
  [PIPELINE_STAGES.UNIT_TEST]: '单元测试',
  [PIPELINE_STAGES.BUILD]: '构建镜像',
  [PIPELINE_STAGES.DEPLOY_FAT]: '部署fat环境',
  [PIPELINE_STAGES.API_TEST]: 'API测试',
  [PIPELINE_STAGES.DEPLOY_STG]: '部署test环境',
};

const PIPELINE_STAGE_ARRAY = map(PIPELINE_STAGES, (value) => value);

const PIPELINE_STAGE_OPTIONS = map(PIPELINE_STAGES, (type) => ({
  value: type,
  label: PIPELINE_STAGE_NAMES[type],
}));

export {
  PIPELINE_STAGES,
  PIPELINE_STAGE_NAMES,
  PIPELINE_STAGE_ARRAY,
  PIPELINE_STAGE_OPTIONS,
};

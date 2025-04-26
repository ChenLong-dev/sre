import { GRPC, RESTFUL } from '@/constants/applicationServiceTypes';
import {
  CRON_JOB,
  ONE_TIME_JOB,
  SERVICE,
  WORKER,
} from '@/constants/applicationTypes';

export const languageConfig = [
  { text: 'GO', value: 'Go' },
  { text: 'NodeJS', value: 'JavaScript' },
  { text: 'Lua', value: 'Lua' },
  { text: 'Python', value: 'Python' },
  { text: 'PHP', value: 'PHP' },
  { text: 'Other', value: 'Others' },
];

export const applicationType = [
  { name: SERVICE, value: SERVICE },
  { name: WORKER, value: WORKER },
  { name: CRON_JOB, value: CRON_JOB },
  { name: ONE_TIME_JOB, value: ONE_TIME_JOB },
];

export const applicationSentryType = [
  { name: '不使用', value: 1 },
  // { name: '自动创建', value: 2 },
  // { name: '使用已存在sentry项目', value: 3 },
];

export const isFav = [{ name: '已收藏', value: true }];

export const ACTIONS = {
  FULL_DEPLOY: 'full_deploy',
  CANARY_DEPLOY: 'canary_deploy',
  FULL_CANARY_DEPLOY: 'full_canary_deploy',
  STOP: 'stop',
  RESTART: 'restart',
  DELETE: 'delete',
  RESUME: 'resume',
  MANUAL_LAUNCH: 'manual_launch',
  UPDATE_HPA: 'update_hpa',
  ENABLE_IN_CLUSTER_DNS: 'enable_in_cluster_dns',
  DISABLE_IN_CLUSTER_DNS: 'disable_in_cluster_dns',
};

export const applicationActions = [
  { name: '全量发布', value: ACTIONS.FULL_DEPLOY },
  { name: '金丝雀发布', value: ACTIONS.CANARY_DEPLOY },
  { name: '停止', value: ACTIONS.STOP },
  { name: '重启', value: ACTIONS.RESTART },
  { name: '删除', value: ACTIONS.DELETE },
  { name: '恢复', value: ACTIONS.RESUME },
];

export const deployActions = [
  { name: '全量部署', value: ACTIONS.FULL_DEPLOY },
  { name: '金丝雀部署', value: ACTIONS.CANARY_DEPLOY },
  { name: '基于金丝雀的全量部署', value: ACTIONS.FULL_CANARY_DEPLOY },
  { name: '停止', value: ACTIONS.STOP },
  { name: '重启', value: ACTIONS.RESTART },
  { name: '恢复', value: ACTIONS.RESUME },
  { name: '删除', value: ACTIONS.DELETE },
  { name: '手动启动', value: ACTIONS.MANUAL_LAUNCH },
  { name: '弹性伸缩', value: ACTIONS.UPDATE_HPA },
];

export const applicationServiveType = [
  { name: RESTFUL, value: RESTFUL },
  { name: GRPC, value: GRPC },
];

// pprof type类型
export const pprofTypeConfig = [
  { name: '内存申请', value: 'allocs' },
  { name: '阻塞', value: 'block' },
  { name: '协程', value: 'goroutine' },
  { name: '堆', value: 'heap' },
  { name: '锁', value: 'mutex' },
  { name: '线程创建', value: 'threadcreate' },
  { name: 'cpu跟踪', value: 'profile' },
  { name: '完整跟踪', value: 'trace' },
];

// pprof action
export const pprofActionConfig = [
  { name: '图片', value: 'svg' },
  { name: '树状表', value: 'tree' },
  { name: 'top', value: 'top' },
  { name: '下载源文件', value: 'download' },
];

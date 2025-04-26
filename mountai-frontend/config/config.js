// https://umijs.org/config/
import { defineConfig } from 'umi';
import defaultSettings from './defaultSettings';
import proxy from './proxy';
const { REACT_APP_ENV, AMS_ENV } = process.env;
// console.log('AMS_ENV==>', AMS_ENV, REACT_APP_ENV);
export default defineConfig({
  nodeModulesTransform: {
    type: 'none',
    exclude: [],
  },
  hash: true,
  antd: {},
  dva: {
    hmr: true,
  },
  locale: {
    // default zh-CN
    default: 'zh-CN',
    // default true, when it is true, will use `navigator.language` overwrite default
    antd: true,
    baseNavigator: true,
  },
  dynamicImport: {
    loading: '@/components/PageLoading/index',
  },
  targets: {
    ie: 11,
  },
  extraBabelPlugins: [
    ["import", { "libraryName": "@qt/react", "camel2DashComponentName": false }, "@qt/react"]
  ],
  // umi routes: https://umijs.org/docs/routing
  routes: [
    {
      path: '/user',
      component: '../layouts/UserLayout',
      routes: [
        {
          name: 'login',
          path: '/user/login',
          component: './user/login',
        },
      ],
    },
    {
      path: '/project/:projectId/',
      component: '../layouts/SecurityLayout',
      routes: [
        {
          path: '/project/:projectId/',
          component: '../layouts/BasicLayout',
          // authority: ['admin', 'user'],
          routes: [
            {
              path: '/project/:projectId/image',
              name: 'docker',
              icon: 'RadarChartOutlined',
              routes: [
                {
                  path: '/project/:projectId/image',
                  name: 'list',
                  icon: 'smile',
                  component: './ProjectInfo/Images', // authority: ['admin'],
                },
              ],
            },
            {
              path: '/project/:projectId/',
              name: 'application',
              icon: 'AppstoreOutlined',
              // component: './ListTableList',
              // authority: ['admin'],
              routes: [
                {
                  path: '/project/:projectId/application',
                  name: 'info',
                  icon: 'smile',
                  component: './ProjectInfo/Application', // authority: ['admin'],
                },
                {
                  path: '/project/:projectId/history',
                  name: 'history',
                  icon: 'smile',
                  component: './ProjectInfo/History', // authority: ['admin'],
                },
              ],
            },
          ],
        },
      ],
    },
    {
      path: '/',
      component: '../layouts/SecurityLayout',
      routes: [
        {
          path: '/',
          component: '../layouts/BasicLayout',
          authority: ['admin', 'user'],
          routes: [
            {
              path: '/',
              component: './Dashboard',
            },
            {
              name: 'projects',
              icon: 'ApartmentOutlined',
              path: '/projects',
              routes: [
                {
                  name: 'list',
                  icon: 'ApartmentOutlined',
                  path: '/projects/list',
                  component: './Project/List',
                },
                {
                  name: 'dynamic',
                  icon: 'smile',
                  path: '/projects/dynamic',
                  component: './Project/Dynamic',
                },
                {
                  name: 'create',
                  hideInMenu: true,
                  path: '/projects/create',
                  component: './Project/Edit',
                },
                {
                  name: 'update',
                  hideInMenu: true,
                  path: '/projects/update/:projectId',
                  component: './Project/Edit',
                },
              ],
            },
            {
              name: 'team',
              icon: 'UserSwitchOutlined',
              path: '/team/',
              //   component: './Team',
              authority: ['admin', 'user'],
              routes: [
                {
                  authority: ['admin', 'user'],
                  path: '/team/list',
                  name: 'list',
                  icon: 'UserSwitchOutlined',
                  component: './Team',
                },
                {
                  authority: ['admin', 'user'],
                  path: '/team/:teamId/image_arg_template',
                  name: 'image-arg-template',
                  hideInMenu: true,
                  component: './Team/ImageArgTemplate',
                },
              ],
            },
            {
              component: './404',
            },
          ],
        },
        {
          component: './404',
        },
      ],
    },
    {
      component: './404',
    },
  ],
  // Theme for antd: https://ant.design/docs/react/customize-theme-cn
  theme: {
    // ...darkTheme,
    'primary-color': defaultSettings.primaryColor,
  },
  // @ts-ignore
  title: false,
  ignoreMomentLocale: true,
  proxy: proxy[REACT_APP_ENV || 'dev'],
  manifest: {
    basePath: '/',
  },
});

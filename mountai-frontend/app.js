const path = require('path');
const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');

let apiServer = 'http://mountai-api.prod.svc.cluster.local';

if (process.env.NODE_ENV === 'test') {
  apiServer = 'http://mountai-api.test.svc.cluster.local';
}

const app = express();
const port = 3001;

console.log('服务器地址:', apiServer);

app.use(express.static(path.join(__dirname, 'dist')));

app.use('/amsapi', createProxyMiddleware({
  target: apiServer,
  changeOrigin: true,
  logLevel: 'debug',
  pathRewrite: {
    '^/amsapi': '', // rewrite path
  },
}));


app.get('/*', function (req, res) {
  res.sendFile(path.join(__dirname, 'dist', 'index.html'));
});



app.listen(port, () => console.log(`Example app listening on port ${port}!`));

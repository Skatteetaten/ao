const express = require('express');
const proxy = require('express-http-proxy');
require('./api/index');

const web = express();

web.use(express.static('public'));
web.use('/api', proxy('localhost:9090', {
  proxyReqPathResolver: function (req) {
    return '/api' + req.url;
  }
}));

web.listen(8080, () => {
  console.log('Listening on port', 8080);
});

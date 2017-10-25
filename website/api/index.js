const express = require('express');

const app = express();

const nodejsPort = process.env.NODEJS_PORT || 9090;
app.listen(nodejsPort, () => {
    console.log('Listening on port', nodejsPort);
});
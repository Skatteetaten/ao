const express = require('express');
const { spawnSync } = require('child_process');
const { resolve } = require('path');

const app = express();
const nodejsPort = process.env.NODEJS_PORT || 9090;

const assets = resolve(__dirname, '../assets');
const ao = resolve(assets, 'ao')

app.get('/api/ao', (req, res) => {
  res.download(ao);
});

app.get('/api/version', (req, res) => {
  const version = spawnSync(ao, ['version', '-o', 'json']);

  if (version.error) {
    console.error(version.error.stack);
    res.status(404).send('ao not found');
    return;
  }

  res.json(JSON.parse(version.stdout.toString()));
});

app.listen(nodejsPort, () => {
  console.log('Listening on port', nodejsPort);
});

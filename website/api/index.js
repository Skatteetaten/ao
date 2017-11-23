// TODO: Denne filen kan slettes når Webleveransepakke støtter kun statiske filer (AOS-2048)
const http = require('http')

const server = http.createServer((req, res) => {
    res.end();
});

server.listen(9090);

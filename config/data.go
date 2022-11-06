package config

const version = "1.0"

var defaultConfig = `
FARMBOTURL: "https://my.farmbot.io/api"
siteTitle: "Farmbot Proxy"
sshKey: ~/.ssh/gitkey
PORT: 
  dev: 8001
  prod: 8000
PORTDASHBOARD: 
  dev: 8001
  prod: 8000
PORTMQTT: 
  dev: 1884
  prod: 1883
PORTWS: 
  dev: 1881
  prod: 1882
update: 86400
TLS: true
DASHBOARD: true
domain: "ubuntu.cseco.co.ke"
scheme: https
documentation: true
contacts:
  support_email: 'brian@cseco.co.ke'
copyright:
  startYear: 2021
  url: 'http://www.cseco.co.ke'
  name: CSECO
analytics: ''
secret: ...
logo: 'https://cseco.co.ke/assets/images/logo.png'
tokenSecret: autoGeneratedReally
database:
  dev:
    databaseType: mariadb
    database:
      force: false
      mysql:
          HOST: localhost
          USER: ''
          PASS: ''
          DBNAME: ''
      mariadb:
          HOST: localhost
          USER: ''
          PASS: ''
          DBNAME: ''
  prod:
    databaseType: mariadb
    database:
      force: false
      mysql:
          HOST: localhost
          USER: ''
          PASS: ''
          DBNAME: ''
      mariadb:
          HOST: localhost
          USER: ''
          PASS: ''
          DBNAME: ''
mailer:
  dev:
    host: 
    port: 587
    secure: true
    auth:
      user: ''
      pass: ''
  prod:
    host: 
    port: 587
    secure: true
    auth:
      user: ''
      pass: ''
        
service:
  prod:
    name: farmbotproxy
    displayname: "Farmbot Proxy"
    description: "Farmbot Proxy"
  dev:
    name: farmbotproxyDev
    displayname: "Farmbot Proxy Dev env"
    description: "Farmbot Proxy in Dev env"
  
`

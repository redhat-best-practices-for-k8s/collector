-- MysqlUsername's and MysqlPassword's values are in tnf-secrets/collector-secrets.json
CREATE USER if not exists 'MysqlUsername'@'%' IDENTIFIED BY 'MysqlPassword';
GRANT ALL PRIVILEGES ON cnf.claim TO 'MysqlUsername'@'%';
GRANT ALL PRIVILEGES ON cnf.claim_result TO 'MysqlUsername'@'%';
GRANT ALL PRIVILEGES ON cnf.authenticator TO 'MysqlUsername'@'%';
FLUSH PRIVILEGES;
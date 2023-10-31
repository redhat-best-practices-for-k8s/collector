CREATE USER if not exists 'MysqlUsername'@'%' IDENTIFIED BY 'MysqlPassword';
GRANT ALL PRIVILEGES ON cnf.claim TO 'MysqlUsername'@'%';
GRANT ALL PRIVILEGES ON cnf.claim_result TO 'MysqlUsername'@'%';
GRANT ALL PRIVILEGES ON cnf.authenticator TO 'MysqlUsername'@'%';
FLUSH PRIVILEGES;
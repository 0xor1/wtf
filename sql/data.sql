DROP DATABASE IF EXISTS data;
CREATE DATABASE data
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_general_ci';
USE data;

DROP USER IF EXISTS 'data'@'%';
CREATE USER 'data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data.* TO 'data'@'%';
GRANT INSERT ON data.* TO 'data'@'%';
GRANT UPDATE ON data.* TO 'data'@'%';
GRANT DELETE ON data.* TO 'data'@'%';
GRANT EXECUTE ON data.* TO 'data'@'%';
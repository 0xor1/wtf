DROP DATABASE IF EXISTS todo_auth;
CREATE DATABASE todo_auth
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_general_ci';
USE todo_auth;

DROP TABLE IF EXISTS auths;
CREATE TABLE auths(
    id BINARY(16) NOT NULL,
	email VARCHAR(250) NOT NULL,
    isActivated BOOLEAN NOT NULL,
    registeredOn DATETIME(3) NOT NULL,
	newEmail VARCHAR(250) NULL,
	activateCode VARCHAR(250) NULL,
	changeEmailCode VARCHAR(250) NULL,
	lastPwdResetOn DATETIME(3) NULL,
	salt   VARBINARY(256) NOT NULL,
	pwd    VARBINARY(256) NOT NULL,
	n      MEDIUMINT UNSIGNED NOT NULL,
	r      MEDIUMINT UNSIGNED NOT NULL,
	p      MEDIUMINT UNSIGNED NOT NULL,
    PRIMARY KEY email (email),
    UNIQUE INDEX id (id),
    INDEX(isActivated, registeredOn)
);

-- cleanup old registrations that have not been activated in a week
SET GLOBAL event_scheduler=ON;
DROP EVENT IF EXISTS regCleanup;
CREATE EVENT regCleanup
ON SCHEDULE EVERY 24 HOUR
STARTS CURRENT_TIMESTAMP + INTERVAL 1 HOUR
DO DELETE FROM auths WHERE isActivated=0 AND registeredOn < DATE_SUB(NOW(), INTERVAL 7 DAY);

DROP USER IF EXISTS 'todo_auth'@'%';
CREATE USER 'todo_auth'@'%' IDENTIFIED BY 'C0-Mm-0n-Auth';
GRANT SELECT ON todo_auth.* TO 'todo_auth'@'%';
GRANT INSERT ON todo_auth.* TO 'todo_auth'@'%';
GRANT UPDATE ON todo_auth.* TO 'todo_auth'@'%';
GRANT DELETE ON todo_auth.* TO 'todo_auth'@'%';
GRANT EXECUTE ON todo_auth.* TO 'todo_auth'@'%';
DROP DATABASE IF EXISTS users;
CREATE DATABASE users
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_general_ci';
USE users;

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

-- each of these tables is optional and can be added / removed depending
-- on the features you want or dont want in your application
-- all the id columns are the users id

-- if you want users to have handles/display names/avatar images
DROP TABLE IF EXISTS socials;
CREATE TABLE socials (
    id BINARY(16) NOT NULL,
    handle VARCHAR(20) NOT NULL,
    alias VARCHAR(50) NOT NULL,
    hasAvatar BOOLEAN NOT NULL,
    PRIMARY KEY id (id),
    UNIQUE INDEX handle (handle),
    FOREIGN KEY (id) REFERENCES auths (id) ON DELETE CASCADE
);

-- if you want users to be able to save arbitrary app config settings as a json blob
DROP TABLE IF EXISTS jins;
CREATE TABLE jins (
    id BINARY(16) NOT NULL,
    val VARCHAR(10000) NOT NULL,
    PRIMARY KEY id (id),
    FOREIGN KEY (id) REFERENCES auths (id) ON DELETE CASCADE
);

-- if you want to have the default fcm managed notifications
DROP TABLE IF EXISTS fcms;
CREATE TABLE fcms (
    id BINARY(16) NOT NULL,
    enabled BOOLEAN NOT NULL,
    PRIMARY KEY id (id),
    FOREIGN KEY (id) REFERENCES auths (id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS fcmTokens;
CREATE TABLE fcmTokens (
    topic VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL,
    id BINARY(16) NOT NULL,
    client BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    PRIMARY KEY (topic, token),
    UNIQUE INDEX (id, client),
    INDEX (id, createdOn),
    INDEX(createdOn),
    FOREIGN KEY (id) REFERENCES auths (id) ON DELETE CASCADE
);
-- cleanup old fcm tokens that were createdOn over 2 days ago
SET GLOBAL event_scheduler=ON;
DROP EVENT IF EXISTS fcmTokenCleanup;
CREATE EVENT fcmTokenCleanup
ON SCHEDULE EVERY 24 HOUR
STARTS CURRENT_TIMESTAMP + INTERVAL 1 HOUR
DO DELETE FROM fcmTokens WHERE createdOn < DATE_SUB(NOW(), INTERVAL 2 DAY);

DROP USER IF EXISTS 'users'@'%';
CREATE USER 'users'@'%' IDENTIFIED BY 'C0-Mm-0n-U5-3r5';
GRANT SELECT ON users.* TO 'users'@'%';
GRANT INSERT ON users.* TO 'users'@'%';
GRANT UPDATE ON users.* TO 'users'@'%';
GRANT DELETE ON users.* TO 'users'@'%';
GRANT EXECUTE ON users.* TO 'users'@'%';
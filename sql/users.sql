DROP DATABASE IF EXISTS users;
CREATE DATABASE users
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_general_ci';
USE users;

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id BINARY(16) NOT NULL,
	email VARCHAR(250) NOT NULL,
    registeredOn DATETIME(3) NOT NULL,
    activatedOn DATETIME(3) NOT NULL,
	newEmail VARCHAR(250) NULL,
	activateCode VARCHAR(250) NULL,
	changeEmailCode VARCHAR(250) NULL,
	lastPwdResetOn DATETIME(3) NULL,
    PRIMARY KEY email (email),
    UNIQUE INDEX id (id),
    INDEX(activatedOn, registeredOn),
    UNIQUE INDEX handle (handle)
);

DROP TABLE IF EXISTS jin;
CREATE TABLE jin (
    user BINARY(16) NOT NULL,
    val VARCHAR(10000) NOT NULL,
    PRIMARY KEY user (user),
    FOREIGN KEY (user) REFERENCES users (id) ON DELETE CASCADE
);

# cleanup old registrations that have not been activated in a week
SET GLOBAL event_scheduler=ON;
DROP EVENT IF EXISTS userRegistrationCleanup;
CREATE EVENT userRegistrationCleanup
ON SCHEDULE EVERY 24 HOUR
STARTS CURRENT_TIMESTAMP + INTERVAL 1 HOUR
DO DELETE FROM users WHERE activatedOn=CAST('0000-00-00 00:00:00.000' AS DATETIME(3)) AND registeredOn < DATE_SUB(NOW(), INTERVAL 7 DAY);

DROP TABLE IF EXISTS fcmTokens;
CREATE TABLE fcmTokens (
    topic VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL,
    user BINARY(16) NOT NULL,
    client BINARY(16) NOT NULL,
    createdOn DATETIME(3) NOT NULL,
    PRIMARY KEY (topic, token),
    UNIQUE INDEX (user, client),
    INDEX (user, createdOn),
    INDEX(createdOn),
    FOREIGN KEY (user) REFERENCES users (id) ON DELETE CASCADE
);

# cleanup old fcm tokens that were createdOn over 2 days ago
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
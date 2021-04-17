DROP DATABASE IF EXISTS data;
CREATE DATABASE data
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_general_ci';
USE data;

DROP TABLE IF EXISTS auths;
CREATE TABLE auths (
    id BINARY(16) NOT NULL,
    isActivated BOOLEAN NOT NULL,
    registeredOn DATETIME(3) NOT NULL,
    PRIMARY KEY id (id),
    INDEX isActivated (isActivated, registeredOn)
);
-- cleanup auths that have not been activated in a week
SET GLOBAL event_scheduler=ON;
DROP EVENT IF EXISTS dataRegCleanup;
CREATE EVENT dataRegCleanup
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
DROP TABLE IF EXISTS jin;
CREATE TABLE jin (
    id BINARY(16) NOT NULL,
    val VARCHAR(10000) NOT NULL,
    PRIMARY KEY id (id),
    FOREIGN KEY (id) REFERENCES auths (id) ON DELETE CASCADE
);

-- if you want to have the default fcm managed notifications
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

DROP USER IF EXISTS 'data'@'%';
CREATE USER 'data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON data.* TO 'data'@'%';
GRANT INSERT ON data.* TO 'data'@'%';
GRANT UPDATE ON data.* TO 'data'@'%';
GRANT DELETE ON data.* TO 'data'@'%';
GRANT EXECUTE ON data.* TO 'data'@'%';
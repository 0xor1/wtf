DROP DATABASE IF EXISTS games_data;
CREATE DATABASE games_data
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_unicode_ci';
USE games_data;

DROP TABLE IF EXISTS games;
CREATE TABLE games (
    id BINARY(16) NOT NULL,
    type VARCHAR(20) NOT NULL,
    updatedOn DATETIME(3) NOT NULL,
    isActive BOOLEAN NOT NULL,
    serialized VARBINARY(5000) NOT NULL,
    PRIMARY KEY id (id, type),
    UNIQUE INDEX isActive (isActive, id, type),
    UNIQUE INDEX updatedOn (updatedOn, id, type)
);

DROP TABLE IF EXISTS players;
CREATE TABLE players (
    id BINARY(16) NOT NULL,
    game BINARY(16) NOT NULL,
    PRIMARY KEY id (id, game),
    UNIQUE INDEX game (game, id),
    FOREIGN KEY (game) REFERENCES games (id) ON DELETE CASCADE
);

DROP USER IF EXISTS 'games_data'@'%';
CREATE USER 'games_data'@'%' IDENTIFIED BY 'C0-Mm-0n-Da-Ta';
GRANT SELECT ON games_data.* TO 'games_data'@'%';
GRANT INSERT ON games_data.* TO 'games_data'@'%';
GRANT UPDATE ON games_data.* TO 'games_data'@'%';
GRANT DELETE ON games_data.* TO 'games_data'@'%';
GRANT EXECUTE ON games_data.* TO 'games_data'@'%';
CREATE DATABASE IF NOT EXISTS dl_platform CHARACTER SET utf8 COLLATE utf8_general_ci;

CREATE USER 'user_user'@'%'    IDENTIFIED BY "{{pwd}}";
CREATE USER 'dataset_user'@'%' IDENTIFIED BY "{{pwd}}";
CREATE USER 'train_user'@'%'   IDENTIFIED BY "{{pwd}}";
CREATE USER 'label_user'@'%'   IDENTIFIED BY "{{pwd}}";
CREATE USER 'model_user'@'%'   IDENTIFIED BY "{{pwd}}";
CREATE USER 'image_user'@'%'   IDENTIFIED BY "{{pwd}}";
CREATE USER 'data_user'@'%'    IDENTIFIED BY "{{pwd}}";
CREATE USER 'cluster_user'@'%' IDENTIFIED BY "{{pwd}}";
CREATE USER 'alarm_user'@'%'   IDENTIFIED BY "{{pwd}}";
CREATE USER 'inference_user'@'%' IDENTIFIED BY "{{pwd}}";

GRANT ALL PRIVILEGES ON *.* TO 'user_user'@'%';
GRANT ALL PRIVILEGES ON *.* TO 'dataset_user'@'%';
GRANT ALL PRIVILEGES ON *.* TO 'train_user'@'%';
GRANT ALL PRIVILEGES ON *.* TO 'label_user'@'%';
GRANT ALL PRIVILEGES ON *.* TO 'model_user'@'%';
GRANT ALL PRIVILEGES ON *.* TO 'image_user'@'%';
GRANT ALL PRIVILEGES ON *.* TO 'data_user'@'%';
GRANT ALL PRIVILEGES ON *.* TO 'cluster_user'@'%';
GRANT ALL PRIVILEGES ON *.* TO 'alarm_user'@'%';
GRANT ALL PRIVILEGES ON *.* TO 'inference_user'@'%';

USE dl_platform;
CREATE TABLE IF NOT EXISTS image_configs(
    id BIGINT AUTO_INCREMENT,
    user_id BIGINT NOT NULL DEFAULT 0,
    group_id BIGINT NOT NULL DEFAULT 0,
    image_name VARCHAR(256),
    image_tag VARCHAR(32),
    image_size DOUBLE NOT NULL,
    harbor_path VARCHAR(256),
    prefabricated tinyint(1) NOT NULL DEFAULT 0,
    image_arch VARCHAR(32) NOT NULL DEFAULT 'noarch',
    status tinyint(1) NOT NULL DEFAULT 0,
    extra_param VARCHAR(256) DEFAULT '',
    create_time DATETIME NOT NULL,
    is_share tinyint(1) NOT NULL DEFAULT 0,
    PRIMARY KEY ( id ),
    UNIQUE (harbor_path)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE DATABASE IF NOT EXISTS `events`;
GRANT ALL PRIVILEGES ON `events`.* to 'events'@'%' IDENTIFIED BY 'events';

DROP TABLE IF EXISTS `events`;
CREATE TABLE `events` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `date` DATE NOT NULL,
  UNIQUE KEY `events_id` (`id`),
  UNIQUE KEY `events_name` (`name`)  
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

INSERT INTO `events` SET name='Swartvlei',date="2022-06-24";
INSERT INTO `events` SET name='Branders',date="2022-07-02";

DROP TABLE IF EXISTS `family_person`;
DROP TABLE IF EXISTS `persons`;
CREATE TABLE `persons` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `nat_id` VARCHAR(40) NOT NULL,
  `last_name` VARCHAR(200) NOT NULL,
  `first_name` VARCHAR(200) NOT NULL,
  `dob` DATE NOT NULL,
  `gender` VARCHAR(1) NOT NULL,
  `phone` VARCHAR(20) DEFAULT NULL,
  `email` VARCHAR(200) DEFAULT NULL,
  `password_hash` VARCHAR(40) DEFAULT NULL,
  `tpw` VARCHAR(40) DEFAULT NULL,
  `tpx` DATETIME DEFAULT NULL,
  UNIQUE KEY `persons_id` (`id`),
  UNIQUE KEY `persons_nat_id` (`nat_id`),
  UNIQUE KEY `persons_phone` (`phone`),
  UNIQUE KEY `persons_email` (`email`),
  KEY `persons_tpw` (`tpw`),
  KEY `persons_name` (`last_name`, `first_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

INSERT INTO `persons` SET `nat_id`="7311185229089",`last_name`="Semmelink",`first_name`="Jan",`dob`="1973-11-18",`gender`="M";
INSERT INTO `persons` SET `nat_id`="7204120026084",`last_name`="Semmelink",`first_name`="Anne-Marie",`dob`="1972-04-12",`gender`="F";
INSERT INTO `persons` SET `nat_id`="0402105666083",`last_name`="Semmelink",`first_name`="Riaan",`dob`="2004-02-10",`gender`="M";
INSERT INTO `persons` SET `nat_id`="0602276334086",`last_name`="Semmelink",`first_name`="Stefan",`dob`="2006-02-27",`gender`="M";
INSERT INTO `persons` SET `nat_id`="0602271356084",`last_name`="Semmelink",`first_name`="Anja",`dob`="2006-02-27",`gender`="F";

CREATE TABLE `family_person` (
  `family_id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `person_id` VARCHAR(40) NOT NULL,
  FOREIGN KEY (`person_id`) REFERENCES `persons`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

INSERT INTO `family_person` SET
  `person_id`=(SELECT `id` FROM `persons` WHERE `nat_id`="7311185229089");

INSERT INTO `family_person` SET
  `family_id`=(SELECT `f`.`family_id` FROM `family_person` AS f 
    JOIN `persons` AS p ON f.person_id=p.id
    WHERE `p`.`nat_id`="7311185229089"),
  `person_id`=(SELECT `id` FROM `persons` WHERE `nat_id`="7204120026084");

INSERT INTO `family_person` SET
  `family_id`=(SELECT `f`.`family_id` FROM `family_person` AS f 
    JOIN `persons` AS p ON f.person_id=p.id
    WHERE `p`.`nat_id`="7311185229089"),
  `person_id`=(SELECT `id` FROM `persons` WHERE `nat_id`="0402105666083");

INSERT INTO `family_person` SET
  `family_id`=(SELECT `f`.`family_id` FROM `family_person` AS f 
    JOIN `persons` AS p ON f.person_id=p.id
    WHERE `p`.`nat_id`="7311185229089"),
  `person_id`=(SELECT `id` FROM `persons` WHERE `nat_id`="0602276334086");

INSERT INTO `family_person` SET
  `family_id`=(SELECT `f`.`family_id` FROM `family_person` AS f 
    JOIN `persons` AS p ON f.person_id=p.id
    WHERE `p`.`nat_id`="7311185229089"),
  `person_id`=(SELECT `id` FROM `persons` WHERE `nat_id`="0602271356084");

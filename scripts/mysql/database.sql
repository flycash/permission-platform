-- create the databases
CREATE
DATABASE IF NOT EXISTS `permission`;

-- create the users for each database
CREATE
USER 'permission'@'%' IDENTIFIED BY 'permission';
GRANT CREATE
, ALTER
, INDEX, LOCK TABLES, REFERENCES,
UPDATE,
DELETE
, DROP
,
SELECT,
INSERT
ON `permission`.* TO 'permission'@'%';

FLUSH
PRIVILEGES;
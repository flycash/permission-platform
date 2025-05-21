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
-- MySQL dump 10.13  Distrib 8.0.29, for Linux (x86_64)
--
-- Host: localhost    Database: permission
-- ------------------------------------------------------
-- Server version	8.0.29

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Current Database: `permission`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `permission` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;

USE `permission`;

--
-- Table structure for table `attribute_definitions`
--

DROP TABLE IF EXISTS `attribute_definitions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `attribute_definitions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `biz_id` bigint DEFAULT NULL COMMENT '和name组成唯一索引，比如说代表订单组的biz_id',
  `name` varchar(255) NOT NULL COMMENT '属性名称',
  `description` text COMMENT '属性描述',
  `data_type` varchar(255) NOT NULL COMMENT '属性数据类型',
  `entity_type` enum('subject','resource','environment') NOT NULL COMMENT '属性所属实体类型',
  `validation_rule` longtext COMMENT '验证规则，正则表达式',
  `ctime` bigint DEFAULT NULL COMMENT '创建时间',
  `utime` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_biz_id_name` (`biz_id`,`name`),
  KEY `idx_entity_type` (`entity_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `attribute_definitions`
--

LOCK TABLES `attribute_definitions` WRITE;
/*!40000 ALTER TABLE `attribute_definitions` DISABLE KEYS */;
/*!40000 ALTER TABLE `attribute_definitions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `business_configs`
--

DROP TABLE IF EXISTS `business_configs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `business_configs` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '''业务ID''',
  `owner_id` bigint DEFAULT NULL COMMENT '''业务方ID''',
  `owner_type` enum('person','organization') DEFAULT NULL COMMENT '''业务方类型：person-个人,organization-组织''',
  `name` varchar(255) NOT NULL COMMENT '''业务名称''',
  `rate_limit` bigint DEFAULT '1000' COMMENT '''每秒最大请求数''',
  `token` text NOT NULL COMMENT '''业务方Token，内部包含bizID''',
  `ctime` bigint DEFAULT NULL,
  `utime` bigint DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `business_configs`
--

LOCK TABLES `business_configs` WRITE;
/*!40000 ALTER TABLE `business_configs` DISABLE KEYS */;
INSERT INTO `business_configs` VALUES (1,999,'organization','权限平台管理后台',3000,'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJiaXpfaWQiOjEsImV4cCI6NDkwMzIxNTgyMywiaWF0IjoxNzQ3NTQyMjIzLCJpc3MiOiJwZXJtaXNzaW9uLXBsYXRmb3JtIn0.f6SgU4RsVJqYoJqYQYhMQhjf_Pb7DTl3WloSyd_pmmQ',1747542223971,1747542223971);
/*!40000 ALTER TABLE `business_configs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `environment_attribute_values`
--

DROP TABLE IF EXISTS `environment_attribute_values`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `environment_attribute_values` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `biz_id` bigint DEFAULT NULL COMMENT '业务ID',
  `attribute_id` bigint NOT NULL COMMENT '属性定义ID',
  `value` text COMMENT '属性值，取决于 data_type',
  `ctime` bigint DEFAULT NULL COMMENT '创建时间',
  `utime` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_biz_attribute` (`biz_id`,`attribute_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `environment_attribute_values`
--

LOCK TABLES `environment_attribute_values` WRITE;
/*!40000 ALTER TABLE `environment_attribute_values` DISABLE KEYS */;
/*!40000 ALTER TABLE `environment_attribute_values` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `permission_policies`
--

DROP TABLE IF EXISTS `permission_policies`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `permission_policies` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `biz_id` bigint DEFAULT NULL COMMENT '业务ID',
  `effect` varchar(50) DEFAULT NULL,
  `permission_id` bigint NOT NULL COMMENT '权限ID',
  `policy_id` bigint NOT NULL COMMENT '策略ID',
  `ctime` bigint DEFAULT NULL COMMENT '创建时间',
  `utime` bigint DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_permission_policy_bizId` (`biz_id`,`permission_id`,`policy_id`),
  KEY `idx_biz_id` (`biz_id`),
  KEY `idx_permission_id` (`permission_id`),
  KEY `idx_policy_id` (`policy_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `permission_policies`
--

LOCK TABLES `permission_policies` WRITE;
/*!40000 ALTER TABLE `permission_policies` DISABLE KEYS */;
/*!40000 ALTER TABLE `permission_policies` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `permissions`
--

DROP TABLE IF EXISTS `permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `permissions` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '''权限ID''',
  `biz_id` bigint NOT NULL COMMENT '''业务ID''',
  `name` varchar(255) NOT NULL COMMENT '''权限名称''',
  `description` text COMMENT '''权限描述''',
  `resource_id` bigint NOT NULL COMMENT '''关联的资源ID，创建后不可修改''',
  `resource_type` varchar(255) NOT NULL COMMENT '''资源类型，冗余字段，加速查询''',
  `resource_key` varchar(255) NOT NULL COMMENT '''资源业务标识符 (如 用户ID, 文档路径)，冗余字段，加速查询''',
  `action` varchar(255) NOT NULL COMMENT '''操作类型''',
  `metadata` text COMMENT '''权限元数据，可扩展字段''',
  `ctime` bigint DEFAULT NULL,
  `utime` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz_resource_action` (`biz_id`,`resource_id`,`action`),
  KEY `idx_biz_action` (`biz_id`,`action`),
  KEY `idx_biz_resource_type` (`biz_id`,`resource_type`),
  KEY `idx_biz_resource_key` (`biz_id`,`resource_key`),
  KEY `idx_resource_id` (`resource_id`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `permissions`
--

LOCK TABLES `permissions` WRITE;
/*!40000 ALTER TABLE `permissions` DISABLE KEYS */;
INSERT INTO `permissions` VALUES (1,1,'business_configs-read','business_configs-read',1,'system_table','/admin/business_configs','read','',1747542224006,1747542224006),(2,1,'business_configs-write','business_configs-write',1,'system_table','/admin/business_configs','write','',1747542224010,1747542224010),(3,1,'resources-read','resources-read',2,'system_table','/admin/resources','read','',1747542224013,1747542224013),(4,1,'resources-write','resources-write',2,'system_table','/admin/resources','write','',1747542224017,1747542224017),(5,1,'permissions-read','permissions-read',3,'system_table','/admin/permissions','read','',1747542224020,1747542224020),(6,1,'permissions-write','permissions-write',3,'system_table','/admin/permissions','write','',1747542224024,1747542224024),(7,1,'roles-read','roles-read',4,'system_table','/admin/roles','read','',1747542224027,1747542224027),(8,1,'roles-write','roles-write',4,'system_table','/admin/roles','write','',1747542224031,1747542224031),(9,1,'role_inclusions-read','role_inclusions-read',5,'system_table','/admin/role_inclusions','read','',1747542224035,1747542224035),(10,1,'role_inclusions-write','role_inclusions-write',5,'system_table','/admin/role_inclusions','write','',1747542224039,1747542224039),(11,1,'role_permissions-read','role_permissions-read',6,'system_table','/admin/role_permissions','read','',1747542224042,1747542224042),(12,1,'role_permissions-write','role_permissions-write',6,'system_table','/admin/role_permissions','write','',1747542224045,1747542224045),(13,1,'user_roles-read','user_roles-read',7,'system_table','/admin/user_roles','read','',1747542224049,1747542224049),(14,1,'user_roles-write','user_roles-write',7,'system_table','/admin/user_roles','write','',1747542224052,1747542224052),(15,1,'user_permissions-read','user_permissions-read',8,'system_table','/admin/user_permissions','read','',1747542224056,1747542224056),(16,1,'user_permissions-write','user_permissions-write',8,'system_table','/admin/user_permissions','write','',1747542224059,1747542224059),(17,1,'account-read','account-read',9,'admin_account','/admin/account','read','',1747542224063,1747542224063),(18,1,'account-write','account-write',9,'admin_account','/admin/account','write','',1747542224066,1747542224066);
/*!40000 ALTER TABLE `permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `policies`
--

DROP TABLE IF EXISTS `policies`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `policies` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `biz_id` bigint DEFAULT NULL COMMENT '业务ID',
  `name` varchar(100) NOT NULL COMMENT '策略名称',
  `description` text COMMENT '策略描述',
  `status` enum('active','inactive') NOT NULL DEFAULT 'active' COMMENT '策略状态',
  `ctime` bigint DEFAULT NULL COMMENT '创建时间',
  `utime` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_biz_name` (`name`),
  KEY `idx_status` (`status`),
  KEY `idx_biz_id` (`biz_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `policies`
--

LOCK TABLES `policies` WRITE;
/*!40000 ALTER TABLE `policies` DISABLE KEYS */;
/*!40000 ALTER TABLE `policies` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `policy_rules`
--

DROP TABLE IF EXISTS `policy_rules`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `policy_rules` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `biz_id` bigint DEFAULT NULL COMMENT '业务ID',
  `policy_id` bigint NOT NULL COMMENT '策略ID',
  `attribute_id` bigint NOT NULL COMMENT '属性定义ID',
  `value` text COMMENT '比较值，取决于类型',
  `left` bigint DEFAULT NULL COMMENT '左规则ID',
  `right` bigint DEFAULT NULL COMMENT '右规则ID',
  `operator` varchar(255) NOT NULL COMMENT '操作符',
  `ctime` bigint DEFAULT NULL COMMENT '创建时间',
  `utime` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_attribute_id` (`attribute_id`),
  KEY `idx_biz_id` (`biz_id`),
  KEY `idx_policy_id` (`policy_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `policy_rules`
--

LOCK TABLES `policy_rules` WRITE;
/*!40000 ALTER TABLE `policy_rules` DISABLE KEYS */;
/*!40000 ALTER TABLE `policy_rules` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `resource_attribute_values`
--

DROP TABLE IF EXISTS `resource_attribute_values`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `resource_attribute_values` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `biz_id` bigint DEFAULT NULL COMMENT 'biz_id + resource_key + attr_id 唯一索引',
  `resource_id` bigint NOT NULL COMMENT '资源ID',
  `attribute_id` bigint NOT NULL COMMENT '属性定义ID',
  `value` text NOT NULL COMMENT '属性值，取决于 data_type',
  `ctime` bigint DEFAULT NULL,
  `utime` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_biz_resource_attr` (`biz_id`,`resource_id`,`attribute_id`),
  KEY `idx_resource_id` (`resource_id`),
  KEY `idx_attribute_id` (`attribute_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `resource_attribute_values`
--

LOCK TABLES `resource_attribute_values` WRITE;
/*!40000 ALTER TABLE `resource_attribute_values` DISABLE KEYS */;
/*!40000 ALTER TABLE `resource_attribute_values` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `resources`
--

DROP TABLE IF EXISTS `resources`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `resources` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '资源ID''',
  `biz_id` bigint NOT NULL COMMENT '''业务ID''',
  `type` varchar(100) NOT NULL COMMENT '''资源类型，被冗余，创建后不允许修改''',
  `key` varchar(255) NOT NULL COMMENT '''资源业务标识符 (如 用户ID, 文档路径)，被冗余，创建后不允许修改''',
  `name` varchar(255) NOT NULL COMMENT '''资源名称''',
  `description` text COMMENT '''资源描述''',
  `metadata` text COMMENT '''资源元数据''',
  `ctime` bigint DEFAULT NULL,
  `utime` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz_type_key` (`biz_id`,`type`,`key`),
  KEY `idx_biz_type` (`biz_id`,`type`),
  KEY `idx_biz_key` (`biz_id`,`key`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `resources`
--

LOCK TABLES `resources` WRITE;
/*!40000 ALTER TABLE `resources` DISABLE KEYS */;
INSERT INTO `resources` VALUES (1,1,'system_table','/admin/business_configs','business_configs','','',1747542223976,1747542223976),(2,1,'system_table','/admin/resources','resources','','',1747542223979,1747542223979),(3,1,'system_table','/admin/permissions','permissions','','',1747542223983,1747542223983),(4,1,'system_table','/admin/roles','roles','','',1747542223986,1747542223986),(5,1,'system_table','/admin/role_inclusions','role_inclusions','','',1747542223989,1747542223989),(6,1,'system_table','/admin/role_permissions','role_permissions','','',1747542223993,1747542223993),(7,1,'system_table','/admin/user_roles','user_roles','','',1747542223996,1747542223996),(8,1,'system_table','/admin/user_permissions','user_permissions','','',1747542223999,1747542223999),(9,1,'admin_account','/admin/account','account','','',1747542224003,1747542224003);
/*!40000 ALTER TABLE `resources` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `role_inclusions`
--

DROP TABLE IF EXISTS `role_inclusions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `role_inclusions` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '角色包含关系ID''',
  `biz_id` bigint NOT NULL COMMENT '''业务ID''',
  `including_role_id` bigint NOT NULL COMMENT '''包含者角色ID（拥有其他角色权限）''',
  `including_role_type` varchar(255) NOT NULL COMMENT '''包含者角色类型（冗余字段，加速查询）''',
  `including_role_name` varchar(255) NOT NULL COMMENT '''包含者角色名称（冗余字段，加速查询）''',
  `included_role_id` bigint NOT NULL COMMENT '''被包含角色ID（权限被包含）''',
  `included_role_type` varchar(255) NOT NULL COMMENT '''被包含角色类型（冗余字段，加速查询）''',
  `included_role_name` varchar(255) NOT NULL COMMENT '''被包含角色名称（冗余字段，加速查询）''',
  `ctime` bigint DEFAULT NULL,
  `utime` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz_including_included` (`biz_id`,`including_role_id`,`included_role_id`),
  KEY `idx_biz_including_role` (`biz_id`,`including_role_id`),
  KEY `idx_biz_included_role` (`biz_id`,`included_role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `role_inclusions`
--

LOCK TABLES `role_inclusions` WRITE;
/*!40000 ALTER TABLE `role_inclusions` DISABLE KEYS */;
/*!40000 ALTER TABLE `role_inclusions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `role_permissions`
--

DROP TABLE IF EXISTS `role_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `role_permissions` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '''角色权限关联关系ID''',
  `biz_id` bigint NOT NULL COMMENT '''业务ID''',
  `role_id` bigint NOT NULL COMMENT '''角色ID''',
  `permission_id` bigint NOT NULL COMMENT '''权限ID''',
  `role_name` varchar(255) NOT NULL COMMENT '''角色名称（冗余字段，加速查询）''',
  `role_type` varchar(255) NOT NULL COMMENT '''角色类型（冗余字段，加速查询）''',
  `resource_type` varchar(255) NOT NULL COMMENT '''资源类型（冗余字段，加速查询）''',
  `resource_key` varchar(255) NOT NULL COMMENT '''资源标识符（冗余字段，加速查询）''',
  `permission_action` varchar(255) NOT NULL COMMENT '''操作类型（冗余字段，加速查询）''',
  `ctime` bigint DEFAULT NULL,
  `utime` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz_role_permission` (`biz_id`,`role_id`,`permission_id`),
  KEY `idx_biz_permission` (`biz_id`,`permission_id`),
  KEY `idx_biz_role_type` (`biz_id`,`role_type`),
  KEY `idx_biz_resource_type` (`biz_id`,`resource_type`),
  KEY `idx_biz_action` (`biz_id`,`permission_action`),
  KEY `idx_biz_resource_key_action` (`biz_id`,`resource_type`,`resource_key`,`permission_action`),
  KEY `idx_biz_role` (`biz_id`,`role_id`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `role_permissions`
--

LOCK TABLES `role_permissions` WRITE;
/*!40000 ALTER TABLE `role_permissions` DISABLE KEYS */;
INSERT INTO `role_permissions` VALUES (1,1,1,1,'权限平台管理后台系统管理员','admin_account','system_table','/admin/business_configs','read',1747542224073,1747542224073),(2,1,1,2,'权限平台管理后台系统管理员','admin_account','system_table','/admin/business_configs','write',1747542224078,1747542224078),(3,1,1,3,'权限平台管理后台系统管理员','admin_account','system_table','/admin/resources','read',1747542224083,1747542224083),(4,1,1,4,'权限平台管理后台系统管理员','admin_account','system_table','/admin/resources','write',1747542224087,1747542224087),(5,1,1,5,'权限平台管理后台系统管理员','admin_account','system_table','/admin/permissions','read',1747542224091,1747542224091),(6,1,1,6,'权限平台管理后台系统管理员','admin_account','system_table','/admin/permissions','write',1747542224095,1747542224095),(7,1,1,7,'权限平台管理后台系统管理员','admin_account','system_table','/admin/roles','read',1747542224099,1747542224099),(8,1,1,8,'权限平台管理后台系统管理员','admin_account','system_table','/admin/roles','write',1747542224103,1747542224103),(9,1,1,9,'权限平台管理后台系统管理员','admin_account','system_table','/admin/role_inclusions','read',1747542224108,1747542224108),(10,1,1,10,'权限平台管理后台系统管理员','admin_account','system_table','/admin/role_inclusions','write',1747542224113,1747542224113),(11,1,1,11,'权限平台管理后台系统管理员','admin_account','system_table','/admin/role_permissions','read',1747542224117,1747542224117),(12,1,1,12,'权限平台管理后台系统管理员','admin_account','system_table','/admin/role_permissions','write',1747542224122,1747542224122),(13,1,1,13,'权限平台管理后台系统管理员','admin_account','system_table','/admin/user_roles','read',1747542224126,1747542224126),(14,1,1,14,'权限平台管理后台系统管理员','admin_account','system_table','/admin/user_roles','write',1747542224131,1747542224131),(15,1,1,15,'权限平台管理后台系统管理员','admin_account','system_table','/admin/user_permissions','read',1747542224135,1747542224135),(16,1,1,16,'权限平台管理后台系统管理员','admin_account','system_table','/admin/user_permissions','write',1747542224140,1747542224140),(17,1,1,17,'权限平台管理后台系统管理员','admin_account','admin_account','/admin/account','read',1747542224146,1747542224146),(18,1,1,18,'权限平台管理后台系统管理员','admin_account','admin_account','/admin/account','write',1747542224160,1747542224160);
/*!40000 ALTER TABLE `role_permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `roles`
--

DROP TABLE IF EXISTS `roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `roles` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '角色ID''',
  `biz_id` bigint NOT NULL COMMENT '''业务ID''',
  `type` varchar(255) NOT NULL COMMENT '''角色类（被冗余，创建后不可修改）''',
  `name` varchar(255) NOT NULL COMMENT '''角色名称（被冗余，创建后不可修改）''',
  `description` text COMMENT '''角色描述''',
  `metadata` text COMMENT '''角色元数据，可扩展字段''',
  `ctime` bigint DEFAULT NULL,
  `utime` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz_type_name` (`biz_id`,`type`,`name`),
  KEY `idx_biz_id` (`biz_id`),
  KEY `idx_role_type` (`type`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `roles`
--

LOCK TABLES `roles` WRITE;
/*!40000 ALTER TABLE `roles` DISABLE KEYS */;
INSERT INTO `roles` VALUES (1,1,'admin_account','权限平台管理后台系统管理员','具有权限平台管理后台内最高管理权限','',1747542224070,1747542224070);
/*!40000 ALTER TABLE `roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `subject_attribute_values`
--

DROP TABLE IF EXISTS `subject_attribute_values`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `subject_attribute_values` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `biz_id` bigint DEFAULT NULL COMMENT 'biz_id + subject_id + attr_id 唯一索引',
  `subject_id` bigint NOT NULL COMMENT '主体ID，通常是用户ID',
  `attribute_id` bigint NOT NULL COMMENT '属性定义ID',
  `value` text NOT NULL COMMENT '属性值，取决于 data_type',
  `ctime` bigint DEFAULT NULL,
  `utime` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_biz_subject_attr` (`biz_id`,`subject_id`,`attribute_id`),
  KEY `idx_subject_id` (`subject_id`),
  KEY `idx_attribute_id` (`attribute_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `subject_attribute_values`
--

LOCK TABLES `subject_attribute_values` WRITE;
/*!40000 ALTER TABLE `subject_attribute_values` DISABLE KEYS */;
/*!40000 ALTER TABLE `subject_attribute_values` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_permissions`
--

DROP TABLE IF EXISTS `user_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_permissions` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '''用户权限关联关系ID''',
  `biz_id` bigint NOT NULL COMMENT '''业务ID''',
  `user_id` bigint NOT NULL COMMENT '''用户ID''',
  `permission_id` bigint NOT NULL COMMENT '''权限ID''',
  `permission_name` varchar(255) NOT NULL COMMENT '''权限名称（冗余字段，加速查询与展示）''',
  `resource_type` varchar(255) NOT NULL COMMENT '''资源类型（冗余字段，加速查询）''',
  `resource_key` varchar(255) NOT NULL COMMENT '''资源标识符（冗余字段，加速查询）''',
  `permission_action` varchar(255) NOT NULL COMMENT '''操作类型（冗余字段，加速查询）''',
  `start_time` bigint NOT NULL COMMENT '''权限生效时间''',
  `end_time` bigint NOT NULL COMMENT '''权限失效时间''',
  `effect` enum('allow','deny') NOT NULL DEFAULT 'allow' COMMENT '''用于额外授予权限，或者取消权限，理论上不应该出现同时allow和deny，出现了就是deny优先于allow''',
  `ctime` bigint DEFAULT NULL,
  `utime` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz_user_permission` (`biz_id`,`user_id`,`permission_id`),
  KEY `idx_biz_user` (`biz_id`,`user_id`),
  KEY `idx_biz_permission` (`biz_id`,`permission_id`),
  KEY `idx_biz_effect` (`biz_id`,`effect`),
  KEY `idx_biz_resource_type` (`biz_id`,`resource_type`),
  KEY `idx_biz_action` (`biz_id`,`permission_action`),
  KEY `idx_time_range` (`biz_id`,`start_time`,`end_time`),
  KEY `idx_current_valid` (`biz_id`,`effect`,`start_time`,`end_time`),
  KEY `idx_biz_resource_key_action` (`biz_id`,`resource_type`,`resource_key`,`permission_action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_permissions`
--

LOCK TABLES `user_permissions` WRITE;
/*!40000 ALTER TABLE `user_permissions` DISABLE KEYS */;
/*!40000 ALTER TABLE `user_permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_roles`
--

DROP TABLE IF EXISTS `user_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_roles` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户角色关联关系主键''',
  `biz_id` bigint NOT NULL COMMENT '''业务ID''',
  `user_id` bigint NOT NULL COMMENT '''用户ID''',
  `role_id` bigint NOT NULL COMMENT '''角色ID''',
  `role_name` varchar(255) NOT NULL COMMENT '''角色名称（冗余字段，加速查询）''',
  `role_type` varchar(255) NOT NULL COMMENT '''角色类型（冗余字段，加速查询）''',
  `start_time` bigint NOT NULL COMMENT '''授予角色生效时间''',
  `end_time` bigint NOT NULL COMMENT '''授予角色失效时间''',
  `ctime` bigint DEFAULT NULL,
  `utime` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz_user_role` (`biz_id`,`user_id`,`role_id`),
  KEY `idx_biz_user` (`biz_id`,`user_id`),
  KEY `idx_biz_role` (`biz_id`,`role_id`),
  KEY `idx_biz_user_role_validity` (`biz_id`,`user_id`,`role_type`,`start_time`,`end_time`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_roles`
--

LOCK TABLES `user_roles` WRITE;
/*!40000 ALTER TABLE `user_roles` DISABLE KEYS */;
INSERT INTO `user_roles` VALUES (1,1,999,1,'权限平台管理后台系统管理员','admin_account',1747542224165,4903215824165,1747542224165,1747542224165);
/*!40000 ALTER TABLE `user_roles` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-05-18  4:23:44

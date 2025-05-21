-- 权限
INSERT INTO `permission`.`permissions` (`id`, `biz_id`, `name`, `description`, `resource_id`, `resource_type`, `resource_key`, `action`, `metadata`, `ctime`, `utime`) VALUES (63, 10000, '代码仓库权限', '代码仓库权限', 0, 'code', 'user.com', 'read', '[1]', 1747834083694, 1747834083694);
-- 权限策略关联
INSERT INTO `permission`.`permission_policies` (`id`, `biz_id`, `effect`, `permission_id`, `policy_id`, `ctime`, `utime`) VALUES (44, 10000, 'allow', 63, 44, 1747834083712, 1747834083712);
-- 策略
INSERT INTO `permission`.`policies` (`id`, `biz_id`, `name`, `description`, `status`, `ctime`, `utime`, `execute_type`) VALUES (44, 10000, '代码仓库读取策略', '允许用户读取代码仓库', 'active', 1747834083697, 1747834083697, 'logic');

-- 策略规则
INSERT INTO `permission`.`policy_rules` (`id`, `biz_id`, `policy_id`, `attr_def_id`, `value`, `left`, `right`, `operator`, `ctime`, `utime`) VALUES (10010, 10000, 44, 10006, '[\"程序员\",\"经理\"]', 0, 0, 'IN', 0, 1747834083698);
INSERT INTO `permission`.`policy_rules` (`id`, `biz_id`, `policy_id`, `attr_def_id`, `value`, `left`, `right`, `operator`, `ctime`, `utime`) VALUES (10011, 10000, 44, 10007, '@day(20:28)', 0, 0, '>=', 0, 1747834083700);
INSERT INTO `permission`.`policy_rules` (`id`, `biz_id`, `policy_id`, `attr_def_id`, `value`, `left`, `right`, `operator`, `ctime`, `utime`) VALUES (10012, 10000, 44, 10007, '@day(22:28)', 0, 0, '<=', 0, 1747834083702);
INSERT INTO `permission`.`policy_rules` (`id`, `biz_id`, `policy_id`, `attr_def_id`, `value`, `left`, `right`, `operator`, `ctime`, `utime`) VALUES (10013, 10000, 44, 10008, '办公室', 0, 0, '=', 0, 1747834083704);
INSERT INTO `permission`.`policy_rules` (`id`, `biz_id`, `policy_id`, `attr_def_id`, `value`, `left`, `right`, `operator`, `ctime`, `utime`) VALUES (10014, 10000, 44, 10009, 'user.com', 0, 0, '=', 0, 1747834083705);
INSERT INTO `permission`.`policy_rules` (`id`, `biz_id`, `policy_id`, `attr_def_id`, `value`, `left`, `right`, `operator`, `ctime`, `utime`) VALUES (10015, 10000, 44, 0, '', 10010, 10011, 'AND', 0, 1747834083707);
INSERT INTO `permission`.`policy_rules` (`id`, `biz_id`, `policy_id`, `attr_def_id`, `value`, `left`, `right`, `operator`, `ctime`, `utime`) VALUES (10016, 10000, 44, 0, '', 10015, 10012, 'AND', 0, 1747834083708);
INSERT INTO `permission`.`policy_rules` (`id`, `biz_id`, `policy_id`, `attr_def_id`, `value`, `left`, `right`, `operator`, `ctime`, `utime`) VALUES (10017, 10000, 44, 0, '', 10016, 10013, 'AND', 0, 1747834083710);
INSERT INTO `permission`.`policy_rules` (`id`, `biz_id`, `policy_id`, `attr_def_id`, `value`, `left`, `right`, `operator`, `ctime`, `utime`) VALUES (10018, 10000, 44, 0, '', 10017, 10014, 'AND', 0, 1747834083711);
-- 资源值
INSERT INTO `permission`.`resource_attribute_values` (`id`, `biz_id`, `resource_id`, `attr_def_id`, `value`, `ctime`, `utime`) VALUES (45, 10000, 33, 10009, 'user.com', 1747834083691, 1747834083691);
-- 资源
INSERT INTO `permission`.`resources` (`id`, `biz_id`, `type`, `key`, `name`, `description`, `metadata`, `ctime`, `utime`) VALUES (33, 10000, 'code', 'user.com', 'user.com', '用户', '[1]', 1747832187665, 1747832187665);
-- 主体值
INSERT INTO `permission`.`subject_attribute_values` (`id`, `biz_id`, `subject_id`, `attr_def_id`, `value`, `ctime`, `utime`) VALUES (45, 10000, 22, 10006, '程序员', 1747834083689, 1747834083689);

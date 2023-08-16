use cnf;

insert into claim (cnf_version, executed_by, partner_name) values ('v4.2.1', 'QE', 'Partner-1');
insert into claim_result (claim_id, suite_name, test_id, test_status)
  (select (select max(id) from claim), 'lifecycle', 'scaling-test-1', 'pass');
insert into claim_result (claim_id, suite_name, test_id, test_status)
  (select (select max(id) from claim), 'lifecycle', 'scaling-test-2', 'skip');
insert into claim_result (claim_id, suite_name, test_id, test_status)
    (select (select max(id) from claim), 'lifecycle', 'scaling-test-3', 'pass');

insert into claim (cnf_version, executed_by, partner_name) values ('v4.2.1', 'QE', 'Partner-2');
insert into claim_result (claim_id, suite_name, test_id, test_status)
  (select (select max(id) from claim), 'affiliated-certification', 'kernel-test', 'fail');
insert into claim_result (claim_id, suite_name, test_id, test_status)
  (select (select max(id) from claim), 'affiliated-certification', 'another-kernel-test', 'fail');
insert into claim_result (claim_id, suite_name, test_id, test_status)
    (select (select max(id) from claim), 'affiliated-certification', 'diff-kernel-test', 'skip');
insert into claim_result (claim_id, suite_name, test_id, test_status)
    (select (select max(id) from claim), 'performance', 'rt-app-isolated', 'pass');

USE cnf;

SET GLOBAL event_scheduler = ON;

CREATE EVENT delete_ci_entries_from_claim_result
ON SCHEDULE EVERY 24 HOUR ENABLE
DO DELETE FROM claim_result
WHERE claim_id IN (
    SELECT id
    FROM claim
    WHERE executed_by = 'CI' AND upload_time < DATE_SUB(NOW(), INTERVAL 1 WEEK)
);

CREATE EVENT delete_ci_entries_from_authenticator
ON SCHEDULE EVERY 1 MINUTE ENABLE
DO DELETE FROM authenticator
WHERE partner_name IN (
    SELECT partner_name
    FROM claim
    WHERE executed_by = 'CI' AND upload_time < DATE_SUB(NOW(), INTERVAL 1 MINUTE)
);

CREATE EVENT delete_ci_entries_from_claim
ON SCHEDULE EVERY 24 HOUR ENABLE
DO DELETE FROM claim
WHERE executed_by = 'CI' AND upload_time < DATE_SUB(NOW(), INTERVAL 1 WEEK);

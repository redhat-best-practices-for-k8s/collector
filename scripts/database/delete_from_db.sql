USE cnf;

SET GLOBAL event_scheduler = ON;

CREATE EVENT delete_from_claim_result
ON SCHEDULE EVERY 24 HOUR ENABLE
DO DELETE FROM claim_result
WHERE claim_id IN (
    SELECT id
    FROM claim
    WHERE upload_time < DATE_SUB(NOW(), INTERVAL 6 MONTH)
);

CREATE EVENT delete_from_claim
ON SCHEDULE EVERY 24 HOUR ENABLE
DO DELETE FROM claim
WHERE upload_time < DATE_SUB(NOW(), INTERVAL 6 MONTH);

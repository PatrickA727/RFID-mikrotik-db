ALTER TABLE sold_items
ALTER COLUMN payment_status TYPE BOOLEAN
USING CASE
    WHEN payment_status IN ('true', '1') THEN true
    ELSE false
END;

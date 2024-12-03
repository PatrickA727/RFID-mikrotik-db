ALTER TABLE items
DROP CONSTRAINT fk_type_ref; -- Drops the foreign key constraint

ALTER TABLE items
DROP COLUMN type_ref;        -- Drops the column

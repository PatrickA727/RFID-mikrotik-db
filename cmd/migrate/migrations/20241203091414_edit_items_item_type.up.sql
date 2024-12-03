ALTER TABLE items
ADD COLUMN type_ref VARCHAR(255);

ALTER TABLE items
ADD CONSTRAINT fk_type_ref
FOREIGN KEY (type_ref) REFERENCES item_type (item_type);
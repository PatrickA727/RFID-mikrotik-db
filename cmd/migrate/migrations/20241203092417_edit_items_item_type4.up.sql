ALTER TABLE items
ADD COLUMN type_ref VARCHAR(255) NOT NULL DEFAULT 'Mikrotik RB750Gr3';

ALTER TABLE items
ADD CONSTRAINT fk_type_ref
FOREIGN KEY (type_ref) REFERENCES item_type (item_type);
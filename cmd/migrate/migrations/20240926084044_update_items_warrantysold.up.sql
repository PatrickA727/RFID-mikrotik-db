ALTER TABLE items
ADD COLUMN warranty VARCHAR(100) DEFAULT 'inactive',
ADD COLUMN sold BOOLEAN DEFAULT false;
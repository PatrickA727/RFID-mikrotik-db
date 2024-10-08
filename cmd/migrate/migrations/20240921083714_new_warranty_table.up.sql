CREATE TABLE IF NOT EXISTS warranty (
    id SERIAL PRIMARY KEY,
    item_id INT NOT NULL UNIQUE,
    purchase_date DATE NOT NULL,
    expiration DATE NOT NULL,
    cust_name VARCHAR(255) NOT NULL,
    cust_email VARCHAR(255) NOT NULL,
    cust_phone VARCHAR(30) NOT NULL,
    createdat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);
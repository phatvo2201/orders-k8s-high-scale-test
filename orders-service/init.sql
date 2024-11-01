CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    category VARCHAR(100) NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);

DO $$
BEGIN
    FOR i IN 1..1000 LOOP
        INSERT INTO products (name, code, category, quantity, price)
        VALUES (concat('Product ', i), concat('CODE-', i), 'Category A', 100000, 19.99);
    END LOOP;
END $$;

-- Create demo tables
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    stock INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES customers(id),
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id),
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL
);

-- Seed customers (10,000 rows)
INSERT INTO customers (email, full_name)
SELECT
    'user' || i || '@example.com',
    'User ' || i
FROM generate_series(1, 10000) AS s(i)
ON CONFLICT DO NOTHING;

-- Seed products (1,000 rows)
INSERT INTO products (name, price, stock)
SELECT
    'Product ' || i,
    (random() * 1000)::numeric(10,2),
    (random() * 1000)::int
FROM generate_series(1, 1000) AS s(i);

-- Seed orders (100,000 rows)
INSERT INTO orders (customer_id, total_amount, status)
SELECT
    (random() * 9999 + 1)::int,
    (random() * 5000)::numeric(10,2),
    CASE (random()*3)::int
        WHEN 0 THEN 'pending'
        WHEN 1 THEN 'completed'
        WHEN 2 THEN 'cancelled'
        ELSE 'shipped'
    END
FROM generate_series(1, 100000) AS s(i);

-- Note: No index on orders.customer_id (intentionally missing for demo)

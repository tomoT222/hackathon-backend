-- Users table
CREATE TABLE IF NOT EXISTS users (
    id CHAR(26) PRIMARY KEY COMMENT 'ULID',
    name VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE
);

-- Items table
CREATE TABLE IF NOT EXISTS items (
    id CHAR(26) PRIMARY KEY COMMENT 'ULID',
    name VARCHAR(100) NOT NULL,
    price INT NOT NULL,
    description TEXT,
    user_id CHAR(26) NOT NULL COMMENT 'Seller ID',
    buyer_id CHAR(26) COMMENT 'Buyer ID',
    status VARCHAR(20) DEFAULT 'on_sale' COMMENT 'on_sale, sold',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (buyer_id) REFERENCES users(id)
);

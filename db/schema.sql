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
    FOREIGN KEY (buyer_id) REFERENCES users(id),
    views_count INT DEFAULT 0,
    ai_negotiation_enabled BOOLEAN DEFAULT FALSE,
    min_price INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Messages table (DM/Comments)
CREATE TABLE IF NOT EXISTS messages (
    id CHAR(26) PRIMARY KEY COMMENT 'ULID',
    item_id CHAR(26) NOT NULL,
    sender_id CHAR(26) NOT NULL,
    content TEXT,
    is_ai_response BOOLEAN DEFAULT FALSE,
    is_approved BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id),
    FOREIGN KEY (sender_id) REFERENCES users(id)
);

-- Negotiation Logs table
CREATE TABLE IF NOT EXISTS negotiation_logs (
    id CHAR(26) PRIMARY KEY COMMENT 'ULID',
    item_id CHAR(26) NOT NULL,
    user_id CHAR(26) NOT NULL COMMENT 'Buyer ID',
    proposed_price INT NOT NULL,
    ai_decision VARCHAR(50) NOT NULL COMMENT 'ACCEPT, REJECT, COUNTER',
    ai_reasoning TEXT,
    log_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

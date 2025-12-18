-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(128) PRIMARY KEY COMMENT 'Firebase UID',
    name VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE
);

-- Items table
CREATE TABLE IF NOT EXISTS items (
    id VARCHAR(128) PRIMARY KEY COMMENT 'ULID',
    name VARCHAR(100) NOT NULL,
    price INT NOT NULL,
    description TEXT,
    user_id VARCHAR(128) NOT NULL COMMENT 'Seller ID',
    buyer_id VARCHAR(128) COMMENT 'Buyer ID',
    status VARCHAR(20) DEFAULT 'on_sale' COMMENT 'on_sale, sold',
    FOREIGN KEY (buyer_id) REFERENCES users(id),
    views_count INT DEFAULT 0,
    ai_negotiation_enabled BOOLEAN DEFAULT FALSE,
    min_price INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    image_url LONGTEXT
);

-- Messages table (DM/Comments)
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(128) PRIMARY KEY COMMENT 'ULID',
    item_id VARCHAR(128) NOT NULL,
    sender_id VARCHAR(128) NOT NULL,
    content TEXT NOT NULL,
    is_ai_response BOOLEAN DEFAULT FALSE,
    is_approved BOOLEAN DEFAULT TRUE,
    suggested_price INT DEFAULT NULL COMMENT 'Price suggested by AI or Buyer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id),
    FOREIGN KEY (sender_id) REFERENCES users(id)
);

-- Negotiation Logs table
CREATE TABLE IF NOT EXISTS negotiation_logs (
    id VARCHAR(128) PRIMARY KEY COMMENT 'ULID',
    item_id VARCHAR(128) NOT NULL,
    user_id VARCHAR(128) NOT NULL COMMENT 'Buyer ID',
    proposed_price INT NOT NULL,
    ai_decision VARCHAR(50) NOT NULL COMMENT 'ACCEPT, REJECT, COUNTER, ANSWER',
    counter_price INT COMMENT 'Counter offer price if any',
    ai_reasoning TEXT,
    log_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

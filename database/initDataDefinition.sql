CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(50) PRIMARY KEY,
    track_number VARCHAR(50),
    entry VARCHAR(5),
    locale VARCHAR(10),
    internal_signature VARCHAR(50),
    customer_id VARCHAR(50),
    delivery_service VARCHAR(50),
    shardkey VARCHAR(50),
    sm_id INTEGER,
    date_created TIMESTAMP,
    oof_shard VARCHAR(10)
    );

CREATE TABLE IF NOT EXISTS delivery(
    order_uid VARCHAR(50),
    name VARCHAR(100),
    phone VARCHAR(14),
    zip VARCHAR(20),
    city VARCHAR(50),
    address VARCHAR(100),
    region VARCHAR(100),
    email VARCHAR(345),
    CONSTRAINT pk_delivery PRIMARY KEY (order_uid),
    CONSTRAINT fk_delivery_belongs_to_order FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
    );

CREATE TABLE IF NOT EXISTS payment (
    order_uid VARCHAR(50),
    transaction VARCHAR(50),
    request_id VARCHAR(50),
    currency VARCHAR(3),
    provider VARCHAR(50),
    amount INTEGER,
    payment_dt INTEGER,
    bank VARCHAR(50),
    delivery_cost INTEGER,
    goods_total INTEGER,
    custom_fee INTEGER,
    CONSTRAINT pk_payment PRIMARY KEY (transaction),
    CONSTRAINT uq_order_uid UNIQUE (order_uid),
    CONSTRAINT fk_payment_belongs_to_order FOREIGN KEY (order_uid) REFERENCES orders(order_uid),
    CONSTRAINT ch_amount_not_negative CHECK ( amount >= 0 ),
    CONSTRAINT ch_delivery_cost_not_negative CHECK ( delivery_cost >= 0 ),
    CONSTRAINT ch_goods_total_not_negative CHECK ( goods_total >= 0 ),
    CONSTRAINT ch_custom_fee_not_negative CHECK ( custom_fee >= 0 )
    );

CREATE TABLE IF NOT EXISTS items (
    order_uid VARCHAR(50),
    chrt_id INTEGER,
    track_number VARCHAR(50),
    price INTEGER,
    rid VARCHAR(50),
    name VARCHAR(50),
    sale INTEGER,
    size VARCHAR(50),
    total_price INTEGER,
    nm_id INTEGER,
    brand VARCHAR(50),
    status INTEGER,
    CONSTRAINT pk_items PRIMARY KEY (rid),
    CONSTRAINT fk_items_belongs_to_order FOREIGN KEY (order_uid) REFERENCES orders(order_uid),
    CONSTRAINT ch_price_not_negative CHECK ( price >= 0 ),
    CONSTRAINT ch_sale_not_negative CHECK ( sale >= 0 ),
    CONSTRAINT ch_total_price_not_negative CHECK ( price >= 0 )
    );
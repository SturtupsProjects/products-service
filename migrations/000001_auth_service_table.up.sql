-- Типы ENUM для указания метода оплаты и типа транзакции
CREATE TYPE payment_method AS ENUM ('uzs', 'usd', 'card');
CREATE TYPE transaction_type AS ENUM ('income', 'expense');


-- Таблица категорий товаров
CREATE TABLE product_categories
(
    id         UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    name       VARCHAR(50) NOT NULL,
    created_by UUID        NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Таблица товаров
CREATE TABLE products
(
    id             UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    category_id    UUID REFERENCES product_categories (id) NOT NULL,
    name           VARCHAR(50)                             NOT NULL,
    bill_format    VARCHAR(5)                              NOT NULL, -- можно заменить на ENUM, если есть ограниченное количество форматов
    incoming_price DECIMAL(10, 2)                          NOT NULL,
    standard_price DECIMAL(10, 2)                          NOT NULL,
    total_count    INT       DEFAULT 0,
    created_by     UUID                                    NOT NULL,
    created_at     TIMESTAMP DEFAULT NOW()
);

-- Таблица продаж
CREATE TABLE sales
(
    id               UUID           DEFAULT gen_random_uuid() PRIMARY KEY,
    client_id        UUID           NOT NULL,
    sold_by          UUID           NOT NULL,
    total_sale_price DECIMAL(10, 2) NOT NULL, -- общая сумма заказа
    payment_method   payment_method DEFAULT 'uzs',
    created_at       TIMESTAMP      DEFAULT NOW()
);

-- Таблица товаров, проданных в рамках конкретной продажи
CREATE TABLE sales_items
(
    id          UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    sale_id     UUID REFERENCES sales (id)    NOT NULL,
    product_id  UUID REFERENCES products (id) NOT NULL,
    quantity    INT  DEFAULT 1                NOT NULL,
    sale_price  DECIMAL(10, 2)                NOT NULL,
    total_price DECIMAL(10, 2)                NOT NULL -- общая цена за конкретный товар в заказе
);

-- Категории для учета денежных потоков
CREATE TABLE cash_category
(
    id   UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

-- Таблица денежных потоков
CREATE TABLE cash_flow
(
    id               UUID           DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id          UUID                               NOT NULL,
    transaction_date TIMESTAMP      DEFAULT NOW(),
    amount           DECIMAL(10, 2)                     NOT NULL,
    transaction_type transaction_type                   NOT NULL,
    category_id      UUID REFERENCES cash_category (id) NOT NULL,
    description      VARCHAR(255),
    payment_method   payment_method DEFAULT 'uzs'
);

-- Таблица долгов
CREATE TABLE debts
(
    id            UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    order_id      UUID REFERENCES sales (id) NOT NULL, -- Привязка долга к заказу
    amount_paid   DECIMAL(10, 2)             NOT NULL, -- Сумма, уже оплаченная
    amount_unpaid DECIMAL(10, 2)             NOT NULL, -- Сумма, остающаяся к оплате
    total_debt    DECIMAL(10, 2)             NOT NULL, -- Общая сумма долга
    next_payment  DATE,
    last_paid_day TIMESTAMP DEFAULT NOW(),
    is_fully_paid BOOLEAN   DEFAULT FALSE,
    recipient_id  UUID                       NOT NULL, -- Кто принял платёж
    created_at    TIMESTAMP DEFAULT NOW()
);

-- Таблица платежей по долгам
CREATE TABLE debt_payments
(
    id           UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    debt_id      UUID REFERENCES debts (id) NOT NULL, -- Привязка к задолженности
    payment_date TIMESTAMP DEFAULT NOW(),
    amount       DECIMAL(10, 2)             NOT NULL, -- Сумма частичного платежа
    paid_by      UUID                       NOT NULL  -- Кто внес платёж
);

-- Таблица закупок
CREATE TABLE purchases
(
    id             UUID           DEFAULT gen_random_uuid() PRIMARY KEY,
    supplier_id    UUID           NOT NULL,      -- Название поставщика или имя компании
    purchased_by   UUID           NOT NULL,      -- Кто произвел закупку
    total_cost     DECIMAL(10, 2) NOT NULL,      -- Общая сумма закупки
    payment_method payment_method DEFAULT 'uzs', -- Способ оплаты
    description    TEXT,
    created_at     TIMESTAMP      DEFAULT NOW()  -- Время создания записи
);

-- Таблица товаров, закупленных в рамках конкретной закупки
CREATE TABLE purchase_items
(
    id             UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    purchase_id    UUID REFERENCES purchases (id) NOT NULL, -- Ссылка на закупку
    product_id     UUID REFERENCES products (id)  NOT NULL, -- Ссылка на товар
    quantity       INT                            NOT NULL, -- Количество закупленного товара
    purchase_price DECIMAL(10, 2)                 NOT NULL, -- Цена закупки за единицу товара
    total_price    DECIMAL(10, 2)                 NOT NULL  -- Общая стоимость конкретного товара в закупке
);

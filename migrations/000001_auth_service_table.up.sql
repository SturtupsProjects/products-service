-- Типы ENUM для указания метода оплаты и типа транзакции
CREATE TYPE payment_method AS ENUM ('uzs', 'usd', 'card');
CREATE TYPE transaction_type AS ENUM ('income', 'expense');


-- Таблица категорий товаров
CREATE TABLE product_categories
(
    id         UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    name       VARCHAR(50)                  NOT NULL,
    image_url  VARCHAR   DEFAULT 'no image' NOT NULL,
    branch_id  UUID                         NOT NULL,
    company_id UUID                         NOT NULL,
    created_by UUID                         NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Таблица товаров
CREATE TABLE products
(
    id             UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    category_id    UUID REFERENCES product_categories (id) NOT NULL,
    name           VARCHAR(150)                            NOT NULL,
    image_url      VARCHAR   DEFAULT 'no image'            NOT NULL,
    bill_format    VARCHAR(5)                              NOT NULL,
    incoming_price DECIMAL(15, 2)                          NOT NULL,
    standard_price DECIMAL(15, 2)                          NOT NULL,
    total_count    INT       DEFAULT 0 CHECK (total_count >= 0),
    branch_id      UUID                                    NOT NULL,
    company_id     UUID                                    NOT NULL,
    created_by     UUID                                    NOT NULL,
    created_at     TIMESTAMP DEFAULT NOW()
);


-- Таблица продаж
CREATE TABLE sales
(
    id               UUID           DEFAULT gen_random_uuid() PRIMARY KEY,
    client_id        UUID           NOT NULL,
    sold_by          UUID           NOT NULL,
    total_sale_price DECIMAL(15, 2) NOT NULL, -- общая сумма заказа
    payment_method   payment_method DEFAULT 'uzs',
    branch_id        UUID           NOT NULL,
    company_id       UUID           NOT NULL,
    created_at       TIMESTAMP      DEFAULT NOW()
);

-- Таблица товаров, проданных в рамках конкретной продажи
CREATE TABLE sales_items
(
    id          UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    sale_id     UUID REFERENCES sales (id)    NOT NULL,
    product_id  UUID REFERENCES products (id) NOT NULL,
    quantity    INT       DEFAULT 1           NOT NULL,
    sale_price  DECIMAL(15, 2)                NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW(),
    branch_id   UUID                          NOT NULL,
    company_id  UUID                          NOT NULL,
    total_price DECIMAL(15, 2)                NOT NULL -- общая цена за конкретный товар в заказе
);

-- Таблица денежных потоков
CREATE TABLE cash_flow
(
    id               UUID           DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id          UUID             NOT NULL,
    transaction_date TIMESTAMP      DEFAULT NOW(),
    amount           DECIMAL(15, 2)   NOT NULL,
    transaction_type transaction_type NOT NULL,
    description      VARCHAR(255),
    branch_id        UUID             NOT NULL,
    company_id       UUID             NOT NULL,
    payment_method   payment_method DEFAULT 'uzs'
);

-- Таблица закупок
CREATE TABLE purchases
(
    id             UUID           DEFAULT gen_random_uuid() PRIMARY KEY,
    supplier_id    UUID                         NOT NULL, -- Название поставщика или имя компании
    purchased_by   UUID                         NOT NULL, -- Кто произвел закупку
    total_cost     DECIMAL(15, 2)               NOT NULL, -- Общая сумма закупки
    payment_method payment_method DEFAULT 'uzs' NOT NULL, -- Способ оплаты
    description    TEXT           DEFAULT ''    NOT NULL,
    branch_id      UUID                         NOT NULL,
    company_id     UUID                         NOT NULL,
    created_at     TIMESTAMP      DEFAULT NOW()           -- Время создания записи
);

-- Таблица товаров, закупленных в рамках конкретной закупки
CREATE TABLE purchase_items
(
    id             UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    purchase_id    UUID REFERENCES purchases (id) NOT NULL, -- Ссылка на закупку
    product_id     UUID REFERENCES products (id)  NOT NULL, -- Ссылка на товар
    quantity       INT                            NOT NULL, -- Количество закупленного товара
    purchase_price DECIMAL(15, 2)                 NOT NULL, -- Цена закупки за единицу товара
    total_price    DECIMAL(15, 2)                 NOT NULL, -- Общая стоимость конкретного товара в закупке
    branch_id      UUID                           NOT NULL,
    company_id     UUID                           NOT NULL
);

CREATE TABLE transfers
(
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    transferred_by UUID NOT NULL,
    from_branch_id UUID NOT NULL,
    to_branch_id UUID NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    company_id UUID NOT NULL
);

CREATE TABLE transfer_products
(
    id             UUID DEFAULT gen_random_uuid(),
    product_transfers_id    UUID REFERENCES transfers (id) NOT NULL,
    product_id     UUID REFERENCES products (id)  NOT NULL,
    quantity       INT                            NOT NULL
);



-- Индексы для таблицы product_categories
CREATE INDEX idx_product_categories_company_branch ON product_categories (company_id, branch_id);
CREATE INDEX idx_product_categories_name ON product_categories (name);

-- Индексы для таблицы products
CREATE INDEX idx_products_company_branch ON products (company_id, branch_id);
CREATE INDEX idx_products_name ON products (name);
CREATE INDEX idx_products_category_id ON products (category_id);

-- Индексы для таблицы sales
CREATE INDEX idx_sales_company_branch ON sales (company_id, branch_id);
CREATE INDEX idx_sales_client_id ON sales (client_id);
CREATE INDEX idx_sales_created_at ON sales (created_at);

-- Индексы для таблицы sales_items
CREATE INDEX idx_sales_items_sale_id ON sales_items (sale_id);
CREATE INDEX idx_sales_items_product_id ON sales_items (product_id);
CREATE INDEX idx_sales_items_company_branch ON sales_items (company_id, branch_id);

-- Индексы для таблицы cash_flow
CREATE INDEX idx_cash_flow_company_branch ON cash_flow (company_id, branch_id);
CREATE INDEX idx_cash_flow_transaction_type ON cash_flow (transaction_type);
CREATE INDEX idx_cash_flow_transaction_date ON cash_flow (transaction_date);

-- Индексы для таблицы purchases
CREATE INDEX idx_purchases_company_branch ON purchases (company_id, branch_id);
CREATE INDEX idx_purchases_supplier_id ON purchases (supplier_id);
CREATE INDEX idx_purchases_created_at ON purchases (created_at);

-- Индексы для таблицы purchase_items
CREATE INDEX idx_purchase_items_purchase_id ON purchase_items (purchase_id);
CREATE INDEX idx_purchase_items_product_id ON purchase_items (product_id);
CREATE INDEX idx_purchase_items_company_branch ON purchase_items (company_id, branch_id);

-- Индексы для таблицы product_transfers
CREATE INDEX idx_product_transfers_company ON transfers (company_id);
CREATE INDEX idx_product_transfers_from_to_branch ON transfers (from_branch_id, to_branch_id);
CREATE INDEX idx_product_transfers_created_at ON transfers (created_at);

-- Индексы для таблицы transfer_products
CREATE INDEX idx_transfer_products_transfer_id ON transfer_products (product_transfers_id);
CREATE INDEX idx_transfer_products_product_id ON transfer_products (product_id);

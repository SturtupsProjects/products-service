-- Типы ENUM для указания метода оплаты и типа транзакции
CREATE TYPE payment_method AS ENUM ('uzs', 'usd', 'card');
CREATE TYPE transaction_type AS ENUM ('income', 'expense');


-- Таблица категорий товаров
CREATE TABLE product_categories
(
    id         UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    name       VARCHAR(50)                  NOT NULL,
    image_url  VARCHAR   DEFAULT 'no image' NOT NULL,
    company_id UUID                         NOT NULL,
    created_by UUID                         NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Таблица товаров
CREATE TABLE products
(
    id             UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    category_id    UUID REFERENCES product_categories (id) NOT NULL,
    name           VARCHAR(50)                             NOT NULL,
    image_url      VARCHAR   DEFAULT 'no image'            NOT NULL,
    bill_format    VARCHAR(5)                              NOT NULL, -- можно заменить на ENUM, если есть ограниченное количество форматов
    incoming_price DECIMAL(10,2)                           NOT NULL,
    standard_price DECIMAL(10,2)                           NOT NULL,
    total_count    INT       DEFAULT 0,
    company_id     UUID                                    NOT NULL,
    created_by     UUID                                    NOT NULL,
    created_at     TIMESTAMP DEFAULT NOW()
);

-- Таблица продаж
CREATE TABLE sales
(
    id               UUID           DEFAULT gen_random_uuid() PRIMARY KEY,
    client_id        UUID   NOT NULL,
    sold_by          UUID   NOT NULL,
    total_sale_price DECIMAL(10,2)  NOT NULL, -- общая сумма заказа
    payment_method   payment_method DEFAULT 'uzs',
    company_id       UUID   NOT NULL,
    created_at       TIMESTAMP      DEFAULT NOW()
);

-- Таблица товаров, проданных в рамках конкретной продажи
CREATE TABLE sales_items
(
    id          UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    sale_id     UUID REFERENCES sales (id)    NOT NULL,
    product_id  UUID REFERENCES products (id) NOT NULL,
    quantity    INT       DEFAULT 1           NOT NULL,
    sale_price  DECIMAL(10,2)                 NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW(),
    company_id  UUID                          NOT NULL,
    total_price DECIMAL(10,2)                 NOT NULL -- общая цена за конкретный товар в заказе
);

-- Таблица денежных потоков
CREATE TABLE cash_flow
(
    id               UUID           DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id          UUID                               NOT NULL,
    transaction_date TIMESTAMP      DEFAULT NOW(),
    amount           DECIMAL(10,2)                     NOT NULL,
    transaction_type transaction_type                   NOT NULL,
    description      VARCHAR(255),
    company_id       UUID                               NOT NULL,
    payment_method   payment_method DEFAULT 'uzs'
);

-- Таблица закупок
CREATE TABLE purchases
(
    id             UUID           DEFAULT gen_random_uuid() PRIMARY KEY,
    supplier_id    UUID                         NOT NULL, -- Название поставщика или имя компании
    purchased_by   UUID                         NOT NULL, -- Кто произвел закупку
    total_cost     DECIMAL(10,2)               NOT NULL, -- Общая сумма закупки
    payment_method payment_method DEFAULT 'uzs' NOT NULL, -- Способ оплаты
    description    TEXT           DEFAULT ''    NOT NULL,
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
    purchase_price DECIMAL(10,2)                 NOT NULL, -- Цена закупки за единицу товара
    total_price    DECIMAL(10,2)                 NOT NULL, -- Общая стоимость конкретного товара в закупке
    company_id     UUID                           NOT NULL
);

-- Индекс на company_id, так как это часто используется для фильтрации
CREATE INDEX idx_product_categories_company_id ON product_categories (company_id);

-- Индексы для связи с другими таблицами
CREATE INDEX idx_products_category_id ON products (category_id);
CREATE INDEX idx_products_company_id ON products (company_id);

-- Индексы для поиска по цене и названию
CREATE INDEX idx_products_incoming_price ON products (incoming_price);
CREATE INDEX idx_products_standard_price ON products (standard_price);

-- Индексы для часто используемых колонок
CREATE INDEX idx_sales_client_id ON sales (client_id);
CREATE INDEX idx_sales_sold_by ON sales (sold_by);
CREATE INDEX idx_sales_company_id ON sales (company_id);

-- Индексы для связи с другими таблицами
CREATE INDEX idx_sales_items_sale_id ON sales_items (sale_id);
CREATE INDEX idx_sales_items_product_id ON sales_items (product_id);
CREATE INDEX idx_sales_items_company_id ON sales_items (company_id);

-- Индексы для использования в фильтрах и выборках
CREATE INDEX idx_cash_flow_user_id ON cash_flow (user_id);
CREATE INDEX idx_cash_flow_company_id ON cash_flow (company_id);
CREATE INDEX idx_cash_flow_transaction_type ON cash_flow (transaction_type);

-- Индексы для поиска по компании и поставщику
CREATE INDEX idx_purchases_supplier_id ON purchases (supplier_id);
CREATE INDEX idx_purchases_purchased_by ON purchases (purchased_by);
CREATE INDEX idx_purchases_company_id ON purchases (company_id);

-- Индексы для поиска по товарам и закупкам
CREATE INDEX idx_purchase_items_purchase_id ON purchase_items (purchase_id);
CREATE INDEX idx_purchase_items_product_id ON purchase_items (product_id);
CREATE INDEX idx_purchase_items_company_id ON purchase_items (company_id);

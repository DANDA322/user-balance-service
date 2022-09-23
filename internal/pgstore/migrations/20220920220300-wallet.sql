-- +migrate Up
CREATE TABLE IF NOT EXISTS wallet
(
    id         bigserial PRIMARY KEY   NOT NULL,
    owner_id   int UNIQUE              NOT NULL,
    balance    float                   NOT NULL,
    created_at timestamp DEFAULT NOW() NOT NULL,
    updated_at timestamp DEFAULT NOW() NOT NULL
    );

CREATE TABLE transaction
(
    id               bigserial PRIMARY KEY         NOT NULL,
    wallet_id        bigint REFERENCES wallet (id),
    amount           float                         NOT NULL,
    target_wallet_id bigint REFERENCES wallet (id),
    comment          text                          NOT NULL,
    timestamp        timestamp DEFAULT NOW()       NOT NULL
);

-- +migrate Down
DROP TABLE transaction CASCADE;
DROP TABLE wallet CASCADE;
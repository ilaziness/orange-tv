--bun:split

ALTER TABLE users
    ADD COLUMN str_id VARCHAR(10) NULL COMMENT '10位数字唯一展示ID',
    ADD COLUMN nickname VARCHAR(15) NOT NULL DEFAULT '' COMMENT '昵称';

--bun:split

UPDATE users SET str_id = LPAD(id, 10, '0') WHERE str_id IS NULL;

--bun:split

ALTER TABLE users
    MODIFY str_id VARCHAR(10) NOT NULL,
    ADD UNIQUE INDEX idx_users_str_id (str_id);


CREATE TABLE accounts (
	"account_id" bigserial NOT NULL UNIQUE,
	"account_limit" bigint NOT NULL,
	"balance" bigint not null,
	CONSTRAINT "accounts_pk" PRIMARY KEY ("account_id")
) WITH (
  OIDS=FALSE
);



CREATE TABLE transactions (
	"id" bigserial NOT NULL UNIQUE,
	"account_id" bigint NOT NULL,
	"value" bigint NOT NULL,
	"description" varchar(10),
  "create_tm" timestamp default current_timestamp,
	CONSTRAINT "transactions_pk" PRIMARY KEY ("id")
) WITH (
  OIDS=FALSE
);

ALTER TABLE transactions ADD CONSTRAINT "transactions_fk0" FOREIGN KEY ("account_id") REFERENCES "accounts"("account_id");
CREATE INDEX idx_account_id on transactions(account_id desc);
insert into accounts (account_limit, balance) values (-100000,0);
insert into accounts (account_limit, balance) values (-80000,0);
insert into accounts (account_limit, balance) values (-1000000,0);
insert into accounts (account_limit, balance) values (-10000000,0);
insert into accounts (account_limit, balance) values (-500000,0);

CREATE OR REPLACE FUNCTION create_debit_transaction(accountId int8, description varchar(10), value int8)
RETURNS TABLE(balance int8, account_limit int8) AS $$
DECLARE 
    balance_val int8;
    account_limit_val int8;
begin
		
    SELECT a.balance, a.account_limit 
    INTO balance_val, account_limit_val
    FROM accounts a 
    WHERE a.account_id = accountId
	for update;
    IF balance_val IS NULL then
        RAISE EXCEPTION 'Account Not Found';
    ELSIF balance_val - value < account_limit_val then
        RAISE EXCEPTION 'Insufficient Funds';
    ELSE
        UPDATE accounts a 
        SET balance = balance_val - value
        WHERE account_id = accountId
        RETURNING a.balance, a.account_limit
        INTO balance_val, account_limit_val;
       insert into transactions (account_id,value,description) values ( accountId,value * -1, description);

      END IF;
    
    RETURN QUERY SELECT balance_val, account_limit_val * -1;
END;
$$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS create_credit_transaction(bigint, character varying, bigint);

CREATE OR REPLACE FUNCTION create_credit_transaction(accountId int8, description varchar(10), value int8)
RETURNS TABLE(balance int8, account_limit int8) AS $$
DECLARE 
    balance_val int8;
    account_limit_val int8;
begin
		
    SELECT a.balance, a.account_limit 
    INTO balance_val, account_limit_val
    FROM accounts a 
    WHERE a.account_id = accountId
	for update;
    IF balance_val IS NULL then
        RAISE EXCEPTION 'Account Not Found';
    ELSE
        UPDATE accounts a 
        SET balance = balance_val + value
        WHERE account_id = accountId
        RETURNING a.balance, a.account_limit
        INTO balance_val, account_limit_val;
       insert into transactions (account_id,value,description) values ( accountId,value, description);

      END IF;
    
    RETURN QUERY SELECT balance_val, account_limit_val * -1;
END;
$$ LANGUAGE plpgsql;


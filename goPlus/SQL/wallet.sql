-- public.wallet definition

-- Drop table

-- DROP TABLE public.wallet;

CREATE TABLE public.wallet(
    userid      int4           NULL,     -- 사용자 ID
    coincode    int4           NULL,     -- Coin Code
	amt         numeric(25,8)  NULL,     -- 잔고
	hold        numeric(25,8)  NULL,     -- 증거금
	address     varchar(120)   NULL,     -- 지갑 주소
	CONSTRAINT wallet_pkey PRIMARY KEY (userid, coincode)
);
COMMENT ON TABLE public.wallet IS '사용자 지갑-잔고';
-- Column comments
COMMENT ON COLUMN public.wallet.userid     IS '사용자 ID';
COMMENT ON COLUMN public.wallet.coincode   IS 'Coin Code';
COMMENT ON COLUMN public.wallet.amt        IS '잔고';
COMMENT ON COLUMN public.wallet.hold       IS '증거금';
COMMENT ON COLUMN public.wallet.address    IS '지갑주소';

INSERT INTO public.wallet (userid,coincode,address,amt, hold) VALUES
     (42280,2,NULL,10000000000.00000000, 0),
     (42280,2965,NULL,10000000000.00000000, 0),
     (42281,2,NULL,10000000000.00000000, 0),
     (42281,2965,NULL,10000000000.00000000, 0),
     (42282,2,NULL,10000000000.00000000, 0),
     (42282,2965,NULL,10000000000.00000000, 0),
     (42283,2,NULL,10000000000.00000000, 0),
     (42283,2965,NULL,10000000000.00000000, 0),
     (42284,2,NULL,10000000000.00000000, 0),
     (42284,2965,NULL,10000000000.00000000, 0),
     (42285,2,NULL,10000000000.00000000, 0),
     (42285,2965,NULL,10000000000.00000000, 0),
     (42286,2,NULL,10000000000.00000000, 0),
     (42286,2965,NULL,10000000000.00000000, 0),
     (42287,2,NULL,10000000000.00000000, 0),
     (42287,2965,NULL,10000000000.00000000, 0),
     (42288,2,NULL,10000000000.00000000, 0),
     (42288,2965,NULL,10000000000.00000000, 0),
     (42289,2,NULL,10000000000.00000000, 0),
     (42289,2965,NULL,10000000000.00000000, 0);


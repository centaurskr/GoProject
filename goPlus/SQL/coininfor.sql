-- public.coininfor definition

-- Drop table

-- DROP TABLE public.coininfor;

CREATE TABLE public.coininfor (
	symbol varchar(20) NOT NULL, -- 코인심볼
	itemcode int4 NOT NULL, -- 종목코드
	coincode int4 NOT NULL, -- COIN Code
	mkcode int4 NOT NULL, -- 시장코드
	CONSTRAINT coininfor_pk PRIMARY KEY (itemcode),
	CONSTRAINT coininfor_unique UNIQUE (symbol)
);
CREATE INDEX coininfor_symbol_idx ON public.coininfor USING btree (symbol, itemcode);
COMMENT ON TABLE public.coininfor IS '코인정보';

-- Column comments

COMMENT ON COLUMN public.coininfor.symbol IS '코인심볼';
COMMENT ON COLUMN public.coininfor.itemcode IS '종목코드';
COMMENT ON COLUMN public.coininfor.coincode IS 'COIN Code';
COMMENT ON COLUMN public.coininfor.mkcode IS '시장코드';

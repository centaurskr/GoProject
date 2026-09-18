-- public.tradetbl definition

-- Drop table

-- DROP TABLE public.tradetbl;

CREATE TABLE public.tradetbl (
	pairkey     varchar(20)    NULL,     -- 체결 Pair key
	maketake    bpchar(1)      NULL,     -- Maker Taker (M,T)
	symbol      varchar(20)    NULL,     -- 종목명
	ordtype     bpchar(1)      NULL,     -- 주문구분 (L:지정가, M:시장가)
	buysell     bpchar(1)      NULL,     -- 매도수 구분 (B:매수, S:매도)
	price       numeric(25, 8) NULL,     -- 체결가격
	qty         numeric(25, 8) NULL,     -- 체결수량
	margin      numeric(25, 8) NULL,     -- 주문증거금(차감용)
	fee         numeric(25, 8) NULL,     -- 체결증거금(차감용) 
	chefee      numeric(25, 8) NULL,     -- 체결 수수료
	feerate     numeric(25, 8) NULL,     -- 체결수수료율(maker or taker)

	ordnum      varchar(20)    NULL,     -- 주문번호 yyyymmdd + 9(13)
	cheseq      int            NULL,     -- 체결번호
	cntnum      varchar(20)    NULL,     -- 상대주문번호
	chetime     timestamp(6)   NULL,     -- 체결시각
	userid      int4           NULL,     -- 사용자 ID
	itemcode    int4           NULL,     -- 종목코드
	coincode    int4           NULL,     -- Coin Code
	mkcode      int4           NULL,     -- 시장코드(결제 통화)
	userfield   varchar(60)    NULL,     -- 사용자정의 필드

	CONSTRAINT tradetbl_pkey PRIMARY KEY (itemcode, ordnum, cheseq)
);
CREATE INDEX tradetbl_pairkey_idx ON public.tradetbl USING btree (pairkey);
CREATE INDEX tradetbl_ordnum_idx ON public.tradetbl USING btree (ordnum);
COMMENT ON TABLE public.tradetbl IS '체결정보';

-- Column comments
COMMENT ON COLUMN public.tradetbl.pairkey    IS '체결 Pair key';
COMMENT ON COLUMN public.tradetbl.maketake   IS 'Maker Taker (M,T)';
COMMENT ON COLUMN public.tradetbl.symbol     IS '종목명';
COMMENT ON COLUMN public.tradetbl.ordtype    IS '주문구분 (L:지정가,M:시장가)';
COMMENT ON COLUMN public.tradetbl.buysell    IS '매도수 구분 (B:매수,S:매도)';
COMMENT ON COLUMN public.tradetbl.price      IS '체결가격';
COMMENT ON COLUMN public.tradetbl.qty        IS '체결수량';
COMMENT ON COLUMN public.tradetbl.margin     IS '주문증거금(차감용)';
COMMENT ON COLUMN public.tradetbl.fee        IS '체결증거금(차감용)';
COMMENT ON COLUMN public.tradetbl.chefee     IS '체결 수수료';
COMMENT ON COLUMN public.tradetbl.feerate    IS '체결수수료율(maker/taker)';
COMMENT ON COLUMN public.tradetbl.ordnum     IS '주문번호 yyyymmdd + 9(13)';
COMMENT ON COLUMN public.tradetbl.cheseq     IS '체결번호';
COMMENT ON COLUMN public.tradetbl.cntnum     IS '상대주문번호';
COMMENT ON COLUMN public.tradetbl.chetime    IS '체결시각 Microsecond';
COMMENT ON COLUMN public.tradetbl.userid     IS '사용자 ID';
COMMENT ON COLUMN public.tradetbl.itemcode   IS '종목코드';
COMMENT ON COLUMN public.tradetbl.coincode   IS 'Coin Code';
COMMENT ON COLUMN public.tradetbl.mkcode     IS '시장코드(결제 통화)';
COMMENT ON COLUMN public.tradetbl.userfield  IS '사용자정의 필드';

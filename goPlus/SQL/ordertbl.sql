-- public.ordertbl definition

-- Drop table

-- DROP TABLE public.ordertbl;

CREATE TABLE public.ordertbl (
	recseq      bigserial      NOT NULL, -- record Serial 번호 자동 부여
	symbol      varchar(20)    NULL,     -- 종목명
	userid      int4           NULL,     -- 사용자 ID
	status      varchar(2)     NULL,     -- 주문상태 (GO:running , CC:취소 CM:체결완료, ER:오류)
	itemcode    int4           NULL,     -- 종목코드
	ordnum      varchar(20)    NULL,     -- 주문번호 yyyymmdd + 9(13)
	ordtype     bpchar(1)      NULL,     -- 주문구분 (1:지정가, 2:시장가)
	buysell     bpchar(1)      NULL,     -- 매도수 구분 (B:매수, S:매도)
	price       numeric(25, 8) NULL,     -- 주문가격
	qty         numeric(25, 8) NULL,     -- 주문수량
	remqty      numeric(25, 8) NULL,     -- 주문잔량(체결후 잔량)
	cheqty      numeric(25, 8) NULL,     -- 체결수량
	canqty      numeric(25, 8) NULL,     -- 취소수량
	margin      numeric(25, 8) NULL,     -- 주문 증거금
	fee         numeric(25, 8) NULL,     -- 체결 증거금
	usertype    int4           NULL,     -- 사용자구분(개인 법인 구분)
	makerate    numeric(10, 8) NULL,     -- Maker 수수료율
	takerate    numeric(10, 8) NULL,     -- Taker 수수료율
	callunit    numeric(10, 8) NULL,     -- 호가단위
	mintrade    numeric(10, 8) NULL,     -- 최소주문 금액
	ip          varchar(30)    NULL,     -- 공인 IP
	uuid        varchar(60)    NULL,     -- UUID, DEVICE ID....
	userfield   varchar(60)    NULL,     -- 사용자정의 필드
	coincode    int4           NULL,     -- Coin Code
	mkcode      int4           NULL,     -- 시장코드(결제 통화)
	cancel      bpchar(1)      NULL,     -- 취소구분 (0:정상 주문, 1:취소주문)
	orgnum      varchar(20)    NULL,     -- 원주문번호
	regnum      varchar(20)    NULL,     -- 접수번호
	regtime     timestamp(6)   NULL,     -- 접수시각
	ordtime     timestamp(6)   NULL,     -- 주문 시각-주문번호 채번 시각
	channel     varchar(8)     NULL,     -- 주문 체널
	exrate      numeric(10,4)  NULL,     -- 환율
	errcode     varchar(12)    NULL,     -- 오류 코드 - status가 ER인경우,정상:0
	CONSTRAINT ordertbl_pkey PRIMARY KEY (itemcode, ordnum)
);
CREATE INDEX ordertbl_recseq_idx ON public.ordertbl USING btree (recseq);
CREATE INDEX ordertbl_userid_idx ON public.ordertbl USING btree (userid);
CREATE INDEX ordertbl_ordnum_idx ON public.ordertbl USING btree (ordnum);
CREATE INDEX ordertbl_regnum_idx ON public.ordertbl USING btree (itemcode, regnum);
COMMENT ON TABLE public.ordertbl IS '주문정보';

-- Column comments
COMMENT ON COLUMN public.ordertbl.recseq      IS 'record Serial 번호 자동 부여';
COMMENT ON COLUMN public.ordertbl.symbol      IS '종목명';
COMMENT ON COLUMN public.ordertbl.channel     IS '주문 체널';
COMMENT ON COLUMN public.ordertbl.itemcode    IS '종목코드';
COMMENT ON COLUMN public.ordertbl.ordnum      IS '주문번호';
COMMENT ON COLUMN public.ordertbl.ordtime     IS '주문 시각-주문번호 채번 시각';
COMMENT ON COLUMN public.ordertbl.remqty      IS '주문잔량(체결후 잔량)';
COMMENT ON COLUMN public.ordertbl.cheqty      IS '체결수량';
COMMENT ON COLUMN public.ordertbl.canqty      IS '취소수량';
COMMENT ON COLUMN public.ordertbl.margin      IS '주문증거금';
COMMENT ON COLUMN public.ordertbl.fee         IS '체결증거금';
COMMENT ON COLUMN public.ordertbl.mkcode      IS '시장코드(결제 통화)';
COMMENT ON COLUMN public.ordertbl.userid      IS '사용자 ID';
COMMENT ON COLUMN public.ordertbl.usertype    IS '사용자구분(개인 법인 구분)';
COMMENT ON COLUMN public.ordertbl.coincode    IS 'Coin Code';
COMMENT ON COLUMN public.ordertbl.ordtype     IS '주문구분 (1:지정가, 2:시장가)';
COMMENT ON COLUMN public.ordertbl.buysell     IS '매도수 구분 (B:매수, S:매도)';
COMMENT ON COLUMN public.ordertbl.price       IS '주문가격';
COMMENT ON COLUMN public.ordertbl.qty         IS '주문수량';
COMMENT ON COLUMN public.ordertbl.makerate    IS 'Maker 수수료율';
COMMENT ON COLUMN public.ordertbl.takerate    IS 'Taker 수수료율';
COMMENT ON COLUMN public.ordertbl.callunit    IS '호가단위';
COMMENT ON COLUMN public.ordertbl.mintrade    IS '최소주문 금액';
COMMENT ON COLUMN public.ordertbl.cancel      IS '취소구분 (0:정상 주문, 1:취소주문)';
COMMENT ON COLUMN public.ordertbl.ip          IS '공인 IP';
COMMENT ON COLUMN public.ordertbl.uuid        IS 'UUID, DEVICE ID....';
COMMENT ON COLUMN public.ordertbl.userfield   IS '사용자정의 필드';
COMMENT ON COLUMN public.ordertbl.orgnum      IS '원주문번호';
COMMENT ON COLUMN public.ordertbl.exrate      IS '환율';
COMMENT ON COLUMN public.ordertbl.regnum      IS '접수번호';
COMMENT ON COLUMN public.ordertbl.regtime     IS '접수시각';
COMMENT ON COLUMN public.ordertbl.status      IS '주문상태 (GO:running , CC:취소 CM:체결완료, ER:오류)';
COMMENT ON COLUMN public.ordertbl.errcode     IS '오류 코드 - status가 ER인경우,정상:0';

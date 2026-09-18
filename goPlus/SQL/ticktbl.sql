-- public.ticktbl definition

-- Drop table
-- Drop TABLE public.ticktbl
CREATE TABLE public.ticktbl (
    chetime     TIMESTAMPTZ    NOT NULL,    -- 체결시각 (마이크로초 해상도)
    itemcode    INT            NOT NULL,    -- 종목코드
    symbol      TEXT           NOT NULL,    -- 종목명
    side        CHAR(1)        NOT NULL,    -- 'B'(매수), 'S'(매도)
    price       NUMERIC(25, 8) NOT NULL,    -- 가격
    qty         NUMERIC(25, 8) NOT NULL,    -- 수량
    seq         BIGINT GENERATED ALWAYS AS IDENTITY, -- 자동 증가 순번 보장
    -- PK에 시간(chetime)이 포함되어야 하이퍼테이블 생성이 가능함.
    CONSTRAINT ticktbl_pkey PRIMARY KEY (itemcode, chetime, seq)
);

-- 하이퍼테이블로 변환
SELECT create_hypertable('public.ticktbl', 'chetime', 
    chunk_time_interval => INTERVAL '3 days');

-- 압축 설정 (종목별로 묶어서 압축해야 효율이 극대화됨)
ALTER TABLE public.ticktbl SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'itemcode',
    timescaledb.compress_orderby = 'chetime DESC, seq DESC'
);

-- 7일이 지난 청크는 자동으로 압축 수행
SELECT add_compression_policy('public.ticktbl', INTERVAL '7 days');

-- 1년(365일)이 지난 데이터는 자동으로 삭제
SELECT add_retention_policy('public.ticktbl', INTERVAL '1 year');

COMMENT ON TABLE public.ticktbl IS '틱정보';

COMMENT ON COLUMN public.ticktbl.chetime     IS  '체결시각 (마이크로초 해상도)';
COMMENT ON COLUMN public.ticktbl.itemcode    IS  '종목코드';
COMMENT ON COLUMN public.ticktbl.symbol      IS  '종목명';
COMMENT ON COLUMN public.ticktbl.side        IS  'B(매수), S(매도)';
COMMENT ON COLUMN public.ticktbl.price       IS  '가격';
COMMENT ON COLUMN public.ticktbl.qty         IS  '수량';
COMMENT ON COLUMN public.ticktbl.seq         IS  '자동 증가 순번보장';

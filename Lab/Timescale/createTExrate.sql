-- 1) TimescaleDB 확장 설치 (superuser로)
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 2) 테스트용 테이블 생성
DROP TABLE IF EXISTS public.TS_exrate;

CREATE TABLE public.TS_exrate (
    dtime     TIMESTAMP   NOT NULL,        -- 시각 (타임존 없음)
    currency  VARCHAR(4)  NOT NULL,        -- 통화 (USDT)
    exrate    FLOAT       NOT NULL,        -- 환율
    timestr   VARCHAR(12)                  -- 분단위 시각 String (YYYYMMDDHH24MI)
        GENERATED ALWAYS AS (
            -- YYYY
            (EXTRACT(YEAR FROM dtime)::INT)::TEXT ||
            -- MM
            LPAD((EXTRACT(MONTH  FROM dtime)::INT)::TEXT, 2, '0') ||
            -- DD
            LPAD((EXTRACT(DAY    FROM dtime)::INT)::TEXT, 2, '0') ||
            -- HH24
            LPAD((EXTRACT(HOUR   FROM dtime)::INT)::TEXT, 2, '0') ||
            -- MI
            LPAD((EXTRACT(MINUTE FROM dtime)::INT)::TEXT, 2, '0')
        ) STORED,
    CONSTRAINT TS_exrate_pk PRIMARY KEY (dtime, currency)
);

COMMENT ON TABLE  public.TS_exrate IS 'TimescaleDB 테스트용 환율 시계열 테이블';
COMMENT ON COLUMN public.TS_exrate.dtime     IS '데이터 시각';
COMMENT ON COLUMN public.TS_exrate.currency  IS '통화(USDT)';
COMMENT ON COLUMN public.TS_exrate.exrate    IS '환율';
COMMENT ON COLUMN public.TS_exrate.timestr   IS '분단위 시각';

-- 3) 하이퍼테이블로 전환 (chunk 간격 : 1일)
SELECT create_hypertable('public.TS_exrate', 
						'dtime', 
						chunk_time_interval => INTERVAL '1 day',
						if_not_exists => TRUE);

-- 압축 설정
ALTER TABLE public.TS_exrate
SET (
    timescaledb.compress,
    timescaledb.compress_orderby  = 'dtime DESC',
    timescaledb.compress_segmentby = 'currency'
);

-- 30일 지난 chunk 자동 압축 정책
SELECT add_compression_policy('public.TS_exrate', INTERVAL '30 days');

-- (옵션) 통화로 자주 조회하면 보조 인덱스 생성
CREATE INDEX IF NOT EXISTS idx_TS_exrate_currency ON public.TS_exrate(currency);
CREATE INDEX IF NOT EXISTS idx_TS_exrate_timestr ON public.TS_exrate(timestr);


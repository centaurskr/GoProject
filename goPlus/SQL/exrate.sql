CREATE TABLE public.exratetbl (
    basetime varchar(12) NOT NULL,  -- 분 단위 기준시각(yyyymmddhhmm)
    currency varchar(4)  NOT NULL,  -- 통화 (USD)
    exrate   float       NOT NULL,  -- 환율
    timestr  varchar(18) NOT NULL,  -- Update 시각 yyyymmddhhmmss
    CONSTRAINT exratetbl_pk PRIMARY KEY (basetime,currency)
);
COMMENT ON TABLE public.exratetbl IS '환율 Table';

COMMENT ON COLUMN public.exratetbl.currency IS '통화 (USD)';
COMMENT ON COLUMN public.exratetbl.basetime IS '기준시각(yyyymm)';
COMMENT ON COLUMN public.exratetbl.exrate   IS '환율';
COMMENT ON COLUMN public.exratetbl.timestr  IS 'UPDATE시각 (yyyymmddhhmmss)';


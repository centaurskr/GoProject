CREATE TABLE public.errtbl (
	errcode   varchar(12)  NOT NULL,    -- 오류코드
	class1    varchar(4)   NOT NULL,    -- 대분류
	class2    varchar(4)   NOT NULL,    -- 중분류
	code      integer      NOT NULL,    -- 순번
	message   varchar(255) NOT NULL,    -- Error Message
	CONSTRAINT errtbl_pk PRIMARY KEY (errcode)
);
COMMENT ON TABLE public.errtbl IS 'Error Table';

-- Column comments

COMMENT ON COLUMN public.errtbl.errcode    IS '오류코드';
COMMENT ON COLUMN public.errtbl.class1   IS '대분류';
COMMENT ON COLUMN public.errtbl.class2   IS '중분류';
COMMENT ON COLUMN public.errtbl.code     IS '순번';
COMMENT ON COLUMN public.errtbl.message  IS 'Error Message';

INSERT INTO public.errtbl (errcode, class1,class2,code,message) VALUES
    ('COM-IF-1', 'COM','IF',1 ,'TR 오류 ORDER'),
    ('COM-IF-2', 'COM','IF',2 ,'거래대상 COIN이 없습니다.'),
    ('COM-IF-3', 'COM','IF',3 ,'거래 체널 오류.'),
    ('COM-IF-4', 'COM','IF',4 ,'UUSER ID가 없습니다.'),
    ('COM-IF-5', 'COM','IF',5 ,'개인 법인 구분이 없습니다.'),
    ('COM-IF-7', 'COM','IF',7 ,'매도수 구분오류.'),
    ('COM-IF-8', 'COM','IF',8 ,'주문 가격이 없습니다'),
    ('COM-IF-9', 'COM','IF',9 ,'주문 수량이 없습니다.'),
    ('COM-IF-10','COM','IF',10,'수수료율이 없습니다.(Maker)'),
    ('COM-IF-11','COM','IF',11,'수수료율이 없습니다.(Taker)'),
    ('COM-IF-12','COM','IF',12,'호가단위가 없습니다.(CALLUNIT)'),
    ('COM-IF-13','COM','IF',13,'최소호가수량이 없습니다.(MINTRADE)'),
    ('COM-IF-14','COM','IF',14,'취소구분이 없습니다.(CANCEL)'),
    ('COM-IF-15','COM','IF',15,'취소주문, 원주문번호가 없습니다.(ORGNUM)'),
    ('COM-IF-16','COM','IF',16,'사용자 IP가 없습니다.'),
    ('COM-IF-17','COM','IF',17,'사용자 UUID가 없습니다.'),
    ('COM-IF-18','COM','IF',18,'시장가 매도가능 수량이 없습니다.'),
    ('COM-IF-19','COM','IF',19,'시장가 매수가능 금액이 없습니다.'),
    ('COM-IF-20','COM','IF',20,'지정가 증거금계산 오류.(using)'),
    ('COM-IF-21','COM','IF',21,'지정가 증거금계산 오류.(fee)'),
    ('COM-IF-22','COM','IF',22,'시장가 증거금계산 오류.(fee)'),
    ('COM-IF-23','COM','IF',23,'시장가 증거금계산 오류.(using)'),
    ('ORD-IF-1', 'ORD','IF',1 ,'주문증거금이 부족합니다.'),
    ('ORD-MA-1', 'ORD','MA',1 ,'원주문 호가가  없습니댜. 원주문 가격을 확인하시오.'),
    ('ORD-MA-2', 'ORD','MA',2 ,'취소를 위한 원주문번호가 없습니다.'),
    ('COM-BO-1', 'COM','BO',1 ,'지원하지 않는 종목입니다.'),
    ('COM-BO-2', 'COM','BO',2 ,'주문정지 상태입니다.'),
    ('COM-BO-3', 'COM','BO',3 ,'매도만 가능합니다.'),
    ('COM-IF-0', 'COM','IF',0 ,'주문 Packet format 오류입니다.'),
    ('ORD-MA-3', 'ORD','MA',3 ,'원주문과 취소주문의 UserId가 다릅니다.'),
    ('COM-IF-6', 'COM','IF',6 ,'주문유형 오류(L:지정가, M:시장가)'),
    ('0'       , '0'  ,'0', 0 ,'정상입니다.'),
    ('SYS-DB-1', 'SYS','DB',1 ,'처리오류입니다.'),
    ('ORD-MA-4', 'ORD','MA',4 ,'시장가 오류. 상대호가가 없습니다.'),
    ('SYS-DB-2', 'SYS','DB',2 ,'처리오류입니다(INS-DB)');
    

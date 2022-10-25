package constants

import (
	"time"

	"github.com/andynikk/gofermart/internal/logger"
)

type Statuses int

const (
	StatusNEW Statuses = iota
	StatusPROCESSING
	StatusINVALID
	StatusPROCESSED

	PortServer      = "localhost:8080"
	PortAcSysServer = "localhost:8000"

	QuerySelectUserWithWhereTemplate = `SELECT 
						* 
					FROM 
						gofermart.users
					WHERE 
						"User" = $1;`

	QuerySelectUserWithPasswordTemplate = `SELECT 
						* 
					FROM 
						gofermart.users
					WHERE 
						"User" = $1 and "Password" = $2;`

	QueryAddUserTemplate = `INSERT INTO 
						gofermart.users ("User", "Password") 
					VALUES
						($1, $2);`

	QueryUserOrdersTemplate = `SELECT 
						* 
					FROM 
						gofermart.orders
					WHERE 
						"User" = $1;`

	QueryOrderWhereNumTemplate = `SELECT
									"userID"
									 , "orderID"
									 , "createdAt"
									 , "startedAt"
									 , "finishedAt"
									 , "failedAt"
									 , CASE WHEN NOT "failedAt" ISNULL THEN 'INVALID' ELSE 'NEW' END AS status
								FROM
									gofermart.orders as orders
								WHERE
										"userID" = $1 and "orderID" = $2;`

	QueryListOrderTemplate = `SELECT	
								 orders."orderID"
								, CASE
									 WHEN NOT orders."failedAt" ISNULL THEN 'INVALID'
									 WHEN NOT orders."finishedAt" ISNULL THEN 'PROCESSED'
									 WHEN NOT orders."startedAt" ISNULL THEN 'PROCESSING'
									 ELSE 'NEW'
								 END AS status
								 , COALESCE(OrderAccrua.accrual, 0) as accrual
							     , CASE
									   WHEN NOT orders."failedAt" ISNULL THEN orders."failedAt"
									   WHEN NOT orders."finishedAt" ISNULL THEN orders."finishedAt"
									   WHEN NOT orders."startedAt" ISNULL THEN orders."startedAt"
									   ELSE orders."createdAt"
								 END AS uploaded_at
							     --, orders."userID"
								 --, orders."createdAt"
								 --, orders."startedAt"
								 --, orders."finishedAt"
								 --, orders."failedAt"	
							FROM
								 gofermart.orders AS orders
							
							LEFT JOIN (SELECT
										   oa."Order",
										   SUM(CASE WHEN oa."TypeAccrual" = 'MINUS' THEN oa."Accrual" * -1 ELSE oa."Accrual" end) AS accrual
									   FROM
										   gofermart.order_accrual AS oa
									   GROUP BY
										   oa."Order") AS OrderAccrua
									  ON orders."orderID" = OrderAccrua."Order"
							
							WHERE
									CASE WHEN $1 = '' THEN true ELSE "userID" = $1 END 
							
								ORDER BY orders."createdAt";`

	QueryUserBalansTemplate = `SELECT
									sum(coalesce(OrderAccrua.total, 0)) AS total
									, sum(coalesce(OrderAccrua.withdrawn, 0)) AS withdrawn
									, sum(coalesce(OrderAccrua."current", 0)) as "current"
								FROM
									gofermart.orders
									
									LEFT JOIN (select
									oa."Order"
									, sum(CASE when oa."TypeAccrual" = 'MINUS' then oa."Accrual" * -1 else oa."Accrual" end) as total
									, sum(CASE when oa."TypeAccrual" = 'MINUS' then oa."Accrual" else 0 end) as withdrawn
									, sum(CASE when oa."TypeAccrual" = 'MINUS' then 0 else oa."Accrual" end) as "current"
									from
									gofermart.order_accrual as oa
									group by
									oa."Order") as OrderAccrua
										ON orders."orderID" = OrderAccrua."Order"
									
								WHERE 
									gofermart.orders."userID" = $1`

	QueryOrderBalansTemplate = `SELECT
									sum(coalesce(OrderAccrua.total, 0)) AS total
									, sum(coalesce(OrderAccrua.withdrawn, 0)) AS withdrawn
									, sum(coalesce(OrderAccrua."current", 0)) as "current"
								FROM
									gofermart.orders
									
									LEFT JOIN (select
									oa."Order"
									, sum(CASE when oa."TypeAccrual" = 'MINUS' then oa."Accrual" * -1 else oa."Accrual" end) as total
									, sum(CASE when oa."TypeAccrual" = 'MINUS' then oa."Accrual" else 0 end) as withdrawn
									, sum(CASE when oa."TypeAccrual" = 'MINUS' then 0 else oa."Accrual" end) as "current"
									from
									gofermart.order_accrual as oa
									group by
									oa."Order") as OrderAccrua
										ON orders."orderID" = OrderAccrua."Order"
								WHERE	
									gofermart.orders."userID" = $1 
									and gofermart.orders."orderID" = $2;`

	QueryAddOrderTemplate = `INSERT INTO 
								gofermart.orders ("userID", "orderID", "createdAt") 
							VALUES
								($1, $2, $3);`
	QueryAddAccrual = `INSERT INTO gofermart.order_accrual("Order", "Accrual", "DateAccrual", "TypeAccrual")
							VALUES ($1, $2, $3, $4);`
	QuerySelectAccrual = `SELECT	
								gofermart.orders."orderID"
								, coalesce(OrderAccrua."DateAccrual", '-infinity'::timestamp) as DateAccrual
								, coalesce(OrderAccrua.withdrawn, 0) AS withdrawn
								, coalesce(OrderAccrua."current", 0) as "current"
							FROM
								gofermart.orders
							
							LEFT JOIN (select
										oa."Order"
										,oa."DateAccrual"
										, CASE when oa."TypeAccrual" = 'MINUS' then oa."Accrual" else 0 END as withdrawn
										, CASE when oa."TypeAccrual" = 'MINUS' then 0 else oa."Accrual" END as "current"
										from	
											gofermart.order_accrual as oa
										where oa."TypeAccrual" = $2) as OrderAccrua
									ON orders."orderID" = OrderAccrua."Order"
							WHERE
								gofermart.orders."userID" = $1
								and not OrderAccrua."DateAccrual" ISNULL;`

	QuerySelectAccrualPLUSS = `SELECT
									*
									FROM
										gofermart.order_accrual as oa	
									WHERE
										oa."Order" = $1
										AND oa."TypeAccrual" = 'PLUS'`

	AccountCookies = "gofermarket_authorization"
)

func (s Statuses) String() string {
	return [...]string{"NEW", "PROCESSING", "INVALID", "PROCESSED"}[s]
}

var HashKey = []byte("taekwondo")
var TimeLiveToken = time.Now().Add(time.Hour * 5).Unix()
var Logger logger.Logger

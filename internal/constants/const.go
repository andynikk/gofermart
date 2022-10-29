package constants

import (
	"time"

	"github.com/andynikk/gofermart/internal/logger"
)

type Statuses int
type Answer int

const (
	StatusNEW Statuses = iota
	StatusPROCESSING
	StatusINVALID
	StatusPROCESSED

	AnswerSuccessfully  Answer = iota //200
	AnswerInvalidFormat               //
	AnswerLoginBusy
	AnswerErrorServer
	AnswerInvalidLoginPassword
	AnswerUserNotAuthenticated
	AnswerAccepted
	AnswerUploadedAnotherUser
	AnswerInvalidOrderNumber
	AnswerInsufficientFunds
	AnswerNoContent
	AnswerConflict
	AnswerTooManyRequests

	PortServer      = "localhost:8080"
	PortAcSysServer = "localhost:8000"
	DemoMode        = false

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
										CASE WHEN $1 = '' THEN true ELSE "userID" = $1 END 
											and "orderID" = $2;`

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
									orders."orderID" as orderID 
									, sum(coalesce(OrderAccrua.total, 0)) AS total
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

								GROUP BY
									orders."orderID"`
	QueryUserOrdes = `SELECT
							o."orderID"
							, 0.00 AS total
							, SUM(CASE WHEN oa."TypeAccrual" = 'MINUS' THEN coalesce(oa."Accrual", 0) ELSE 0 END) AS withdrawn
							, SUM(CASE WHEN oa."TypeAccrual" = 'MINUS' THEN 0 ELSE coalesce(oa."Accrual", 0) END) AS "current"
						FROM
							gofermart.orders AS o
						LEFT JOIN gofermart.order_accrual AS oa
							ON o."orderID" = oa."Order"
						WHERE
							o."userID" = $1
						GROUP BY
							o."orderID"
							`

	QueryOrderBalansTemplate = `SELECT
									'' as orderID
									, sum(coalesce(OrderAccrua.total, 0)) AS total
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
									and gofermart.orders."orderID" = $2

								GROUP BY
									orderID;`

	QueryAddOrderTemplate = `INSERT INTO 
								gofermart.orders ("userID", "orderID", "createdAt") 
							VALUES
								($1, $2, $3);`
	QueryAddAccrual = `INSERT INTO gofermart.order_accrual("Accrual", "DateAccrual", "TypeAccrual", "Order")
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

	QueryUpdateStartedAt = `UPDATE gofermart.orders
								SET "startedAt" = $1
								WHERE "orderID" = $2;`
)

func (s Statuses) String() string {
	return [...]string{"NEW", "PROCESSING", "INVALID", "PROCESSED"}[s]
}

var HashKey = []byte("taekwondo")
var TimeLiveToken = time.Now().Add(time.Hour * 5).Unix()
var Logger logger.Logger

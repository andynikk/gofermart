package constants

import (
	"gofermart/internal/logger"
	"time"
)

type Statuses int

const (
	StatusNEW Statuses = iota
	StatusPROCESSING
	StatusINVALID
	StatusPROCESSED

	PortServer = ":8080"

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
	QueryOrderWhereNumTemplate = `SELECT "User", "Order"
					FROM 
						gofermart.orders
					WHERE 
						"Order" = $1;`
	QueryListOrderWhereTemplate = `SELECT 
										gofermart.orders."Order" 
										, CurrentStatus.Status 
										, coalesce(OrderAccrua.accrual, 0) AS Accrual 
										, CurrentStatus.UploadedAt as UploadedAt	
									FROM 
										gofermart.orders
									LEFT JOIN (SELECT 
										DateCurrentStatus."Order", 
										DateCurrentStatus.DateStatus as UploadedAt, 
										gofermart.order_statuses."Status" as Status
											FROM (SELECT 
													"Order" as "Order", 
													Max("DateStatus") as DateStatus
													FROM 
														gofermart.order_statuses		
													GROUP BY 
														gofermart.order_statuses."Order") as DateCurrentStatus	
											LEFT JOIN 
												gofermart.order_statuses 
											ON DateCurrentStatus."Order" = gofermart.order_statuses."Order"
											AND DateCurrentStatus.DateStatus = gofermart.order_statuses."DateStatus") AS CurrentStatus 
									ON gofermart.orders."Order" = CurrentStatus."Order"	
									
									LEFT JOIN (SELECT 
													gofermart.order_statuses."Order", 
													gofermart.order_statuses."DateStatus" as DateCreate
												FROM 
													gofermart.order_statuses
												WHERE 
													gofermart.order_statuses."Status" = 'NEW') as CreateStatus 
										ON orders."Order" = CreateStatus."Order"
									
									LEFT JOIN (select 
													oa."Order", 
													sum(CASE when oa."TypeAccrual" = 'MINUS' then oa."Accrual" * -1 else oa."Accrual" end) as accrual
												from 
													gofermart.order_accrual as oa
												group by 
													oa."Order") as OrderAccrua
										ON orders."Order" = OrderAccrua."Order"				
									
									WHERE gofermart.orders."User" = $1
											ORDER BY CreateStatus.DateCreate;`

	QueryListOrderTemplate = `SELECT 
									gofermart.orders."Order" 
									, CurrentStatus.Status 
									, coalesce(OrderAccrua.accrual, 0) AS Accrual 
									, CurrentStatus.UploadedAt as UploadedAt	
								FROM 
									gofermart.orders
								LEFT JOIN (SELECT 
									DateCurrentStatus."Order", 
									DateCurrentStatus.DateStatus as UploadedAt, 
									gofermart.order_statuses."Status" as Status
										FROM (SELECT 
												"Order" as "Order", 
												Max("DateStatus") as DateStatus
												FROM 
													gofermart.order_statuses		
												GROUP BY 
													gofermart.order_statuses."Order") as DateCurrentStatus	
										LEFT JOIN 
											gofermart.order_statuses 
										ON DateCurrentStatus."Order" = gofermart.order_statuses."Order"
										AND DateCurrentStatus.DateStatus = gofermart.order_statuses."DateStatus") AS CurrentStatus 
								ON gofermart.orders."Order" = CurrentStatus."Order"	
								
								LEFT JOIN (SELECT 
												gofermart.order_statuses."Order", 
												gofermart.order_statuses."DateStatus" as DateCreate
											FROM 
												gofermart.order_statuses
											WHERE 
												gofermart.order_statuses."Status" = 'NEW') as CreateStatus 
									ON orders."Order" = CreateStatus."Order"
								
								LEFT JOIN (select 
												oa."Order", 
												sum(CASE when oa."TypeAccrual" = 'MINUS' then oa."Accrual" * -1 else oa."Accrual" end) as accrual
											from 
												gofermart.order_accrual as oa
											group by 
												oa."Order") as OrderAccrua
									ON orders."Order" = OrderAccrua."Order"				
									
									ORDER BY CreateStatus.DateCreate;`

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
										ON orders."Order" = OrderAccrua."Order"
									
								WHERE 
									gofermart.orders."User" = $1`

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
										ON orders."Order" = OrderAccrua."Order"
								WHERE	
									gofermart.orders."User" = $1 
									and gofermart.orders."Order" = $2;`

	QueryAddOrderTemplate = `INSERT INTO 
						gofermart.orders ("User", "Order") 
					VALUES
						($1, $2);`
	QueryAddStatusTemplate = `INSERT INTO gofermart.order_statuses(
						"Order", "Status", "DateStasus")
						VALUES ($1, $2, $3);`
	QueryAddAccrual = `INSERT INTO gofermart.order_accrual("Order", "Accrual", "DateAccrual", "TypeAccrual")
							VALUES ($1, $2, $3, $4);`
	QuerySelectAccrual = `SELECT	
								gofermart.orders."Order"
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
									ON orders."Order" = OrderAccrua."Order"
							WHERE
								gofermart.orders."User" = $1
								and not OrderAccrua."DateAccrual" ISNULL;`

	//where CASE when $3 = '' then true else oa."TypeAccrual" = $3 END) as OrderAccrua
	//CASE when $1 = '' then true else gofermart.orders."User" = $1 END
	//and CASE when $2 = 0 then true else gofermart.orders."Order" = $2 END;
	AccountCookies = "gofermarket_authorization"
)

func (s Statuses) String() string {
	return [...]string{"NEW", "PROCESSING", "INVALID", "PROCESSED"}[s]
}

var HashKey = []byte("taekwondo")
var TimeLiveToken = time.Now().Add(time.Minute * 5).Unix()
var Logger logger.Logger

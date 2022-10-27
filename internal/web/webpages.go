package web

import (
	"fmt"
)

func StartPage(host string) string {
	content := fmt.Sprintf(`<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<h2>Гофермарт</h2>

				<p>&nbsp;</p>
				
				<p><a href="http://%s/docs/user/register" target="_self">регистрация пользователя</a></p>
				
				<p><a href="http://%s/docs/user/login" target="_self">аутентификация пользователя</a></p>
				
				<p><a href="http://%s/docs/user/order" target="_self">загрузка пользователем номера 
						заказа для расчёта</a></p>
				
				<p><a href="http://%s/docs/user/orders" target="_self">получение списка загруженных 
						пользователем номеров заказов, статусов их обработки и информации о начислениях</a></p>
				
				<p><a href="http://%s/docs/user/balance" target="_self">получение текущего баланса счёта
						баллов лояльности пользователя</a></p>
				
				<p><a href="http://%s/docs/user/balance/withdraw" target="_self">запрос на списание баллов 
						с накопительного счёта в счёт оплаты нового заказа</a></p>
				
				<p><a href="http://%s/docs/user/balance/withdrawals" target="_self">получение информации о 
						выводе средств с накопительного счёта пользователем</a></p>

				<p><a href="http://%s/docs/user/accrual" target="_self">Получить информацию о начислении баллов лояльности 
						по заказу</a></p>
				</body>
				</html>`, host, host, host, host, host, host, host, host)

	return content
}

func RegisterPage(host string) string {
	content := fmt.Sprintf(`<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<h3>Гофермарт</h3>
				
				<p>Регистрация</p>
				
				<p>Имя:&nbsp;<input name="name" id="name" type="text" /></p>
				
				<p>Пароль:&nbsp;<input name="password" id="password" type="password" /></p>
				
				<p><input name="register" type="button" value="Зарегестрировать" onclick="functionToExecute()" /></p>
				<script type="text/javascript">
					async function functionToExecute() {	
						var txtName = document.querySelector("#name");
						var txtPassword = document.querySelector("#password");	
						
						let user = {
						  "login": txtName.value,
							"password": txtPassword.value
						};
						
						let response = await fetch('http://%s/api/user/register', {
						  method: 'POST',
						  headers: {
							'Content-Type': 'application/json;charset=utf-8'
						  },
						  body: JSON.stringify(user)
						});
							
						const token = response.headers.get('Authorization');
//debugger
						localStorage.setItem("token", token);	
						if (token != '') {
							window.location.href = 'http://%s/';
						} else {
							alert(response.status)
						}
					}
				</script>
				</body>
				</html>`, host, host)

	return content
}

func LoginPage(host string) string {
	content := fmt.Sprintf(`<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<h3>Гофермарт</h3>
				
				<p>Аутентификация</p>
				
				<p>Имя:&nbsp;<input name="name" id="name" type="text" /></p>
				
				<p>Пароль:&nbsp;<input name="password" id="password" type="password" /></p>
				
				<p><input name="register" type="button" value="Аутентификация" onclick="functionToExecute()" /></p>

				<script type="text/javascript">
					async function functionToExecute() {	
						var txtName = document.querySelector("#name");
						var txtPassword = document.querySelector("#password");	
						
						let user = {
						  "login": txtName.value,
							"password": txtPassword.value
						};
						
						let response = await fetch('http://%s/api/user/login', {
						  method: 'POST',
						  headers: {
							'Content-Type': 'application/json;charset=utf-8'
						  },
						  body: JSON.stringify(user)
						});
							
						const token = response.headers.get('Authorization');
						localStorage.setItem("token", token);	
						if (token != '') {
							window.location.href = 'http://%s';
						} else {
							alert(response.status)
							window.location.href = 'http://%s/user/register';
						}
					}
				</script>
				</body>
				</html>`, host, host, host)

	return content
}

func OrderPage(host string) string {

	content := fmt.Sprintf(`<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<h3>Гофермарт</h3>
				
				<p>Сделать заказ</p>
				
				<p>Ордер:&nbsp;<input name="order" id="order" type="text" /></p>
						
				<p><input name="register" type="button" value="Заказать" onclick="functionToExecute()" /></p>

				<script type="text/javascript">
					async function functionToExecute() {	
						const txtOrder = document.querySelector("#order");

						let response = await fetch('http://%s/api/user/orders', {
						  method: 'POST',
						  headers: {
							'Content-Type': 'application/text;charset=utf-8',
 							'Authorization': localStorage.getItem("token"),
						  },
						  body: txtOrder.value
						});	

						const httpStatus = response.status
						if (httpStatus == 202) {
							window.location.href = 'http://%s';
						} else {
							alert(httpStatus)
						}
					}
				</script>
				</body>
				</html>`, host, host)

	return content
}

func OrdersPage(host string) string {
	content := fmt.Sprintf(`<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<body onload="loadPage()">
				<h3>Гофермарт</h3>
					
				<pre id='txtOrders'></pre>
				<script type="text/javascript">
					async function loadPage() {

						let response = await fetch('http://%s/api/user/orders', {
						  method: 'GET',
						  headers: {	
 							'Authorization': localStorage.getItem("token"),
						  }
						});	

						const httpStatus = response.status
						let json = await response.json();
						document.getElementById("txtOrders").innerHTML = "Заказы (статус ответа " + httpStatus + "): <br />" 
																	+ JSON.stringify(json, null, 4);		
					}
				</script>
				</body>
				</html>`, host)

	return content
}

func BalancePage(host string) string {
	content := fmt.Sprintf(`<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<body onload="loadPage()">
				<h3>Гофермарт</h3>
					
				<pre id='txtOrders'></pre>
				<script type="text/javascript">
					async function loadPage() {

						let response = await fetch('http://%s/api/user/balance', {
						  method: 'GET',
						  headers: {	
 							'Authorization': localStorage.getItem("token"),
						  }
						});	

						const httpStatus = response.status
						let json = await response.json();
						document.getElementById("txtOrders").innerHTML = "Баланс (статус ответа " + httpStatus + "): <br />" + 
																JSON.stringify(json, null, 4);	

						//if (httpStatus == 200) {
						//	window.location.href = 'http://%s';
						//} else {
						//	alert(httpStatus)
						//}
					}
				</script>
				</body>
				</html>`, host, host)

	return content
}

func BalanceWithdrawPage(host string) string {
	content := fmt.Sprintf(`<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<h3>Гофермарт</h3>
				
				<p>Вывод средств</p>
				
				<p>Заказ:&nbsp;<input name="order" id="order" type="text" /></p>
				
				<p>Баллы спсания:&nbsp;<input name="withdraw" id="withdraw" type="number" /></p>
				
				<p><input name="register" type="button" value="Вывести" onclick="functionToExecute()" /></p>

				<script type="text/javascript">
					async function functionToExecute() {	
						var txtOrder = document.querySelector("#order");
						var txtWithdraw = document.querySelector("#withdraw");	
						
						let user = {
						  "order": txtOrder.value,
							"sum": txtWithdraw.valueAsNumber
						};
						
						let response = await fetch('http://%s/api/user/balance/withdraw', {
						  method: 'POST',
						  headers: {
							'Content-Type': 'application/json;charset=utf-8',
							'Authorization': localStorage.getItem("token"),
						  },
						  body: JSON.stringify(user)
						});
									
						const httpStatus = response.status
						if (httpStatus == 200) {
							window.location.href = 'http://%s';
						} else {
							alert(httpStatus)
						}
					}
				</script>
				</body>
				</html>`, host, host)

	return content
}

func BalanceWithdrawsPage(host string) string {
	content := fmt.Sprintf(`<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<body onload="loadPage()">
				<h3>Гофермарт</h3>
					
				<pre id='txtOrders'></pre>
				<script type="text/javascript">
					async function loadPage() {		
						
						let response = await fetch('http://%s/api/user/withdrawals', {
						  method: 'GET',
						  headers: {	
 							'Authorization': localStorage.getItem("token"),
						  }
						});
									
						const httpStatus = response.status
						let json = await response.json();
						document.getElementById("txtOrders").innerHTML = "Вывод средств (статус ответа " + httpStatus + "): <br />" 
																	+ JSON.stringify(json, null, 4);
					}
				</script>
				</body>
				</html>`, host)

	return content
}

func AccrualPage(host string) string {

	content := fmt.Sprintf(`<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<h3>Гофермарт</h3>
				
				<p>Начисление балов</p>
				
				<p>Ордер:&nbsp;<input name="order" id="order" type="text" /></p>
						
				<p><input name="register" type="button" value="Получить баллы лояльности" onclick="functionToExecute()" /></p>
				<p><br /></p>
				<pre id='txtOrders'></pre>
				<script type="text/javascript">
					async function functionToExecute() {	
						const txtOrder = document.querySelector("#order");

						let response = await fetch('http://%s/api/orders/' + txtOrder.value, {
						  method: 'GET',
						  headers: {
							'Content-Type': 'application/text;charset=utf-8',
 							'Authorization': localStorage.getItem("token"),
						  },
						});	

						const httpStatus = response.status
						if (httpStatus == 200) {
							window.location.href = 'http://%s';
						} else {
							alert(httpStatus)
						}
					}
				</script>
				</body>
				</html>`, host, host)

	return content
}

package web

func StartPage() string {
	content := `<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<h2>Гофермарт</h2>

				<p>&nbsp;</p>
				
				<p><a href="http://localhost:8080/user/register" target="_blank">регистрация пользователя</a></p>
				
				<p><a href="http://localhost:8080/user/login" target="_blank">аутентификация пользователя</a></p>
				
				<p><a href="http://localhost:8080/api/user/orders" target="_blank">загрузка пользователем номера 
						заказа для расчёта</a></p>
				
				<p><a href="http://localhost:8080/api/user/orders" target="_blank">получение списка загруженных 
						пользователем номеров заказов, статусов их обработки и информации о начислениях</a></p>
				
				<p><a href="http://localhost:8080/api/user/balance" target="_blank">получение текущего баланса счёта
						баллов лояльности пользователя</a></p>
				
				<p><a href="http://localhost:8080/api/user/balance/withdraw" target="_blank">запрос на списание баллов 
						с накопительного счёта в счёт оплаты нового заказа</a></p>
				
				<p><a href="http://localhost:8080/api/user/balance/withdrawals" target="_blank">получение информации о 
						выводе средств с накопительного счёта пользователем</a></p>
				</body>
				</html>`

	return content
}

func OrderPage(arrOrder []string) string {
	content := `<!DOCTYPE html>
				<html>
				<head>
  					<meta charset="UTF-8">
  					<title>МЕТРИКИ</title>
				</head>
				<body>
				<h1>МЕТРИКИ</h1>
				<ul>
				`

	for _, val := range arrOrder {
		content = content + `<li><b>` + val + `</b></li>` + "\n"
	}
	content = content + `</ul>
						</body>
						</html>`

	return content
}

func LoginPage() string {
	content := `<!DOCTYPE html>
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
				
				<p><input name="register" type="button" value="Sign In" onclick="functionToExecute()" /></p>

				<script type="text/javascript">
					function functionToExecute() {
						const xhr = new XMLHttpRequest();
						var txtName = document.querySelector("#name");
						var txtPassword = document.querySelector("#password");
						const json = {
							"name": txtName.value,
							"password": txtPassword.value
						};
						xhr.open('POST', 'http://localhost:8080/api/user/login');
						xhr.setRequestHeader("Content-Type", "application/json");
						xhr.send(JSON.stringify(json));
					}
				</script>
				</body>
				</html>`

	return content
}

func RegisterPage() string {
	content := `<!DOCTYPE html>
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
					function functionToExecute() {
						const xhr = new XMLHttpRequest();
						var txtName = document.querySelector("#name");
						var txtPassword = document.querySelector("#password");
						const json = {
							"name": txtName.value,
							"password": txtPassword.value
						};
						xhr.open('POST', 'http://localhost:8080/api/user/register');
						xhr.setRequestHeader("Content-Type", "application/json");
						xhr.send(JSON.stringify(json));
					}
				</script>
				</body>
				</html>`

	return content
}

func AuthorizationPage(user string, arrOrders []string) string {
	content := `<!DOCTYPE html>
				<html>
				<head>
  					<meta charset="UTF-8">
  					<title>Г О Ф Е Р М А Р К Е Т</title>
				</head>
				<body>
				<h3>Гофермарт</h3>
				
				<p>` + user + `</p>	
				<ul>
				`
	for _, val := range arrOrders {
		content = content + `<li><b>` + val + `</b></li>` + "\n"
	}
	content = content + `</ul>
						</body>
						</html>`
	return content
}

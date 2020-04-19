#Requisites
- go 1.14+
- postgres
#### Instalar go:
- Linux https://dl.google.com/go/go1.14.1.windows-amd64.msi 
- Mac OS https://dl.google.com/go/go1.14.1.darwin-amd64.pkg
- Windows https://dl.google.com/go/go1.14.1.darwin-amd64.pkg

###Configurar e instalar Postgresql:
https://www.thegeekstuff.com/2009/04/linux-postgresql-install-and-configure-from-source/

#Instalar
Clone este repositorio:``git clone http://wankar.com:3000/kaypi/kaypi_back_geo.git``

Ejecute:

    $cd kaypi_back_geo/
    $go build .  
    
Ejecute `go help buildmode` para que vea los diferentes tipos de build que se pueden hacer. 


#Correr las pruebas
Las pruebas estan en  `test/location_test.go` . Para correr las pruebas la aplicacion debe estar corriendo, ademas es necesario que la base de datos de admins no se halla modificado.

Use `go test [test_file]`.

#Correr la aplicacion
Usted puede correr la aplicacion sin hacer build con el comando `go run .`.
##Manualmente
Exporta ls variables de entorno para la coneccion con la base de datos:

    # nombre del usuario de postgres (es necesario que el usuario tenga
    # permisos para la base de datos. Esto lo puede hacer con `GRANT ALL PRIVILEGES ON
    DATABASE {nombre de la base de datos} TO {usuario};`).
     export      DB_USER="ExName".
    # contraseña del usuario en postgres.
     export      DB_PASS="ExPass"
    # nombre de la base de datos.
     export      DB_NAME="ExUserName"
    # host de la base de datos.
     export      DB_HOST="localhost"
    # Si tiene el SSL habilitado  ("enable" o "disable").
     export      DB_SSL_MODE="disable"
    # endpoint para GraphQl Ex( /location )
     export      ENDPOINT="/location"
    # Direccion del servidor Ex(localhost:8080)
     export      SERVER_ADDRESS=":8080"
     # Tiempo de expiracion del RefreshToken (en horas)
     export     REFRESH_TOKEN_EXP="360h"
     # Tiempo de expiracion del TempToken (en horas o minutos)
     export     TEMP_TOKEN_EXP="15m"
     # Secreto para el JWT (Json Web Token)
     export     SECRET="XXXXX"
	 
Corra el siguiente comando para ejecutar la aplicacion:	 ``go run ./``
##Usando el script run.sh.
Abre el script `run.sh` y configura las variables de entorno  corre `sh script/run.sh`.

#Rest
Todos los endpoints de rest son manejados por el admin.

#####LoginAdmin:
***URI***: /admin/login   
***Method***: POST  
***Descripcion***: Logea un admin en la base de datos con basic authentication. (La base de datos contiene un admin 
con userName: root y password: 12345678. Se recomienda cambiar el password antes de poner el servicio de cara al mundo).    
***Ex***: 
 ```bash
curl -X POST http://localhost:8080/admin/login\
 -H 'Authorization: Basic cm9vdDoxMjM0NTY3OA=='
``` 

***Return***:   
-httpStatusOK(200)      
```json
{"temp_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODcyNzg4NTgsImlhdCI6MTU4NzI3Nzk1OCwiaWQiOiIzIiwicm9sZSI6IkFkbWluIn0.30gn2pK_fIRynj39BCTs7RzDTXv8a32P_FXNPyk9-x8"}
```    
or
```json
{"error": "Error msg", "http_status": 400}
```    
```json
{"error": "Error msg", "http_status": 401}
```    
```json
{"error": "Error msg", "http_status": 500}
```    
#####ChangeAdminPass:
***URI***: /admin/changePassword   
***Method***: POST  
***Args***:     

	last_password: String.  
	new_password: String.   

***Descripcion***: Cambia la contraseña del admin. (Se recomienda cambiar la contraseña periodicamente).
 Requiere un tempToken de admin.    
***Ex***: 
 ```bash
 curl -X POST http://localhost:8080/admin/changePassword \
 -H 'Content-Type: application/json' \
 -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODcyODAxNTksImlhdCI6MTU4NzI3OTI1OSwiaWQiOiIzIiwicm9sZSI6IkFkbWluIn0.KIVKHAnHyS_aWiV3_WMhuiplhIDbCWT_u1S_JEIFQRA' \
 -d '{"last_password": "87654321","new_password": "12345678"}'
``` 

***Return***:     
httpStatusOK(200)   
 or     
```json
{"error": "Error msg", "http_status": 400}
```    
```json
{"error": "Error msg", "http_status": 500}
```    
```json
{"error": "Error msg", "http_status": 511}
```    

#####RegisterUser:
***URI***: /admin/registerUser  
***Method***: POST  
***Args***:  

	"lat": Float,
	"long": Float,
	"category": String
	
***Descripcion***: Registra un usuario en la base de datos y retorna su id con su refreshToken.(la categoria del usuario debe ser "CLIENT" o "SERVICE_PROVIDER", de no ser asi retornara un error).
 Requiere un tempToken de admin.    
***Ex***: 
 ```bash
 curl -X POST http://localhost:8080/admin/registerUser\
 -H 'Content-Type: application/json' \
 -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODcyODg3MDAsImlhdCI6MTU4NzI4NzgwMCwiaWQiOiIzIiwicm9sZSI6IkFkbWluIn0.tkRXRChNJSXB5tRDd-yIHXOM8hfuTRqPOu0WsnX5sKE' \
 -d '{"lat": 1.34122,"long": -5.3, "category":"SERVICE_PROVIDER"}'
``` 

***Return***:     
httpStatusOK(200)   
```json
{
  "user_id":29,
  "refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODg1NzgwMDAsImlhdCI6MTU4NzI4MjAwMCwiaWQiOiIyOSIsInJvbGUiOiJVc2VyIiwidHlwZSI6IlJlZlRva2VuIn0.9RaQQkf-1N7IrM07llAGD_mwF0MF90eWnGqykxI-3sw"
}
```
 or     
```json
{"error": "Error msg", "http_status": 400}
```    
```json
{"error": "Error msg", "http_status": 500}
```    
```json
{"error": "Error msg", "http_status": 511}
```    

#####GetRefreshTokenFromClient:
***URI***: /admin/getRefreshToken
***Method***: POST  
***Args***:  

	"id": Int
***Descripcion***: Obtiene un refreshToken de un usuario dado su id. Si el refreshToken del usuario expiro regresa uno nuevo.
 Requiere un tempToken de admin.    
***Ex***: 
 ```bash
 curl -X POST http://localhost:8080/admin/getRefreshToken\
 -H 'Content-Type: application/json' \
 -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODcyODg3MDAsImlhdCI6MTU4NzI4NzgwMCwiaWQiOiIzIiwicm9sZSI6IkFkbWluIn0.tkRXRChNJSXB5tRDd-yIHXOM8hfuTRqPOu0WsnX5sKE' \
 -d '{"id": 29}'
``` 

***Return***:     
httpStatusOK(200)   
```json
{
"refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODg1NzgwMDAsImlhdCI6MTU4NzI4MjAwMCwiaWQiOiIyOSIsInJvbGUiOiJVc2VyIiwidHlwZSI6IlJlZlRva2VuIn0.9RaQQkf-1N7IrM07llAGD_mwF0MF90eWnGqykxI-3sw"
}
```
 or     
```json
{"error": "Error msg", "http_status": 400}
```    
```json
{"error": "Error msg", "http_status": 500}
```    
```json
{"error": "Error msg", "http_status": 511}
```    

#####GetRefreshTokenFromClient:
***URI***: /admin/deleteUser
***Method***: POST  
***Args***:  

	"id": Int
***Descripcion***: Elimina un usuario de la base de datos dada una id. 
 Requiere un tempToken de admin.    
***Ex***: 
 ```bash
 curl -X POST http://localhost:8080/admin/deleteUser\
 -H 'Content-Type: application/json' \
 -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODcyODIxMTksImlhdCI6MTU4NzI4MTIxOSwiaWQiOiIzIiwicm9sZSI6IkFkbWluIn0.W6JGvPFrLxk1PHxddwhVXO27yfnspQYQ6xLvh-UCVNE' \
 -d '{"id": 29}'
``` 

***Return***:     
httpStatusOK(200)   
 or     
```json
{"error": "Error msg", "http_status": 400}
```    
```json
{"error": "Error msg", "http_status": 500}
```    
```json
{"error": "Error msg", "http_status": 511}
```    

#GraphQl
##Tipos
***User***: 
```
type User{ // User representa un usuario del sistema. 
      id: ID // id del usuario
      category: category // categoria del usuario 
      geo_cord: geo_cord // localizacion del usario (latitud y longitud)
}
```

***geo_cord***:
```
type geo_cord { // geo_cord representa una punto de localizacion. 
    latitude: Float // latitud de la localizacion
    longitude: Float // longitud de la localizacion
}
```

***category***:
```
enum category{ // category representa un la categoria de los usarios.
    CLIENT // representa un usuario cliente
    SERVICE_PROVIDER // representa un provedor de sevicio del sistema
    GENERIC // incluye ambos anteriores (solo usado para algunas queries como allUser(category)). 
     // Usuarios no pueden ser de tipo GENERIC,
     // declarar un usuario con categoria GENERIC puede retornar un error.
}
```

##Query
#####AllUsers:  
***Descripcion***:
 Retorna todos los usuarios que coincidan con una categoria.(Use la categoria GENERIC para obtener todos los usuarios.)
Requiere autenticacion con token en el header.( Con TempToken de usuario, si utiliza un TempToken de admin o un RefreshToken devolvera un error). 

    allUsers(
        category: category
      ): [User]
 
 
***Using Altair:*** 
![alt text](./QueryScreenshots/allUsers.png)

***Using Playground***:
![alt text](./QueryScreenshots/allUsers(Playground).png)

#####User: 
***Descripcion***:
Retorna un usuario que coincida con una id.
Requiere autenticacion con token en el header.( Con TempToken de usuario, si utiliza un TempToken de admin o un RefreshToken devolvera un error). 

    user( 
        id: Int 
      ): User
      
***Using Altair:*** 
![alt text](./QueryScreenshots/user.png)

***Using Playground***:
![alt text](./QueryScreenshots/user(Playground).png)
      
#####CurrentUser:
***Descripcion***:
 Retorna los datos del usuario actual.
Requiere autenticacion con token en el header.( Con TempToken de usuario, si utiliza un TempToken de admin o un RefreshToken devolvera un error). 
 
     currentUser: User
      
***Using Altair:*** 
![alt text](./QueryScreenshots/currentUser.png)

***Using Playground***:
![alt text](./QueryScreenshots/currentUser(Playground).png)
      
##### GetCloseUsers:
***Descripcion***:
 retorna una lista de los usuarios cercanos a una poscicion dada de acuerdo a una resolucion de indice H3. Usa `category=GENERIC` para obtener usuarios de categoria 
`CLIENT` y `SERVICE_PROVIDER`. Para mas informacion sobre las resoluciones de indices H3 consulte: https://h3geo.org/#/documentation/core-library/resolution-table",
Requiere autenticacion con token en el header.( Con TempToken de usuario, si utiliza un TempToken de admin o un RefreshToken devolvera un error). 

    getCloseUsers( 
        origin_lat: Float // latitud del punto del cual se obtendran los usuarios cercanos.
        originLong: Float // longitud del punto del cual se obtendran los usuarios cercanos.
        resolution: Int  // resolucion de indice H3. Valores posibles (0 - 15)
        category: category  // categoria de Usuarios.
      ): [User]

***Using Altair:*** 
![alt text](./QueryScreenshots/getCloseUsers.png)

***Using Playground***:
![alt text](./QueryScreenshots/getCloseUsers(Playground).png)
      
##Mutation
#####UpdateGeoCord:
***Descripcion***:
Actualiza la localizacion de un usuario.
Requiere autenticacion con token en el header.( Con TempToken de usuario, si utiliza un TempToken de admin o un RefreshToken devolvera un error). 

    updateGeoCord(
        newLat: Float
        newLong: Float
      ): User 
      
***Using Altair:*** 
![alt text](./QueryScreenshots/updateGeoCord.png)

***Using Playground***:
![alt text](./QueryScreenshots/updateGeoCord(Playground).png)
      
#####GetUserTempToken:
***Descripcion***:
Obtiene un token temporal a partir de un refresh token.

    getUserTempToken(
        refreshToken: String 
      ): String 
      
***Using Altair:*** 
![alt text](./QueryScreenshots/getUserTempToken.png)

***Using Playground***:
![alt text](./QueryScreenshots/getUserTempToken(Playground).png)
      
##Subscriptions
#####GetUserPos:
***Descripcion***:
Subscribete a un usuario dado su id para obtener los cambios en su posicion.
Requiere autenticacion con token en el header.( Con TempToken de usuario, si utiliza un TempToken de admin o un RefreshToken devolvera un error). 

    getUserPost(id:Int) geo_cord

![alt text](./QueryScreenshots/getUserPos.gif)

#Examples
GraphQl: Mira en la carpeta QueryScreenShots. Hay un archivo .jpg por cada query, mutation y subscriptions.
Rest: puedes ver los test en test/location_test.go. 
  

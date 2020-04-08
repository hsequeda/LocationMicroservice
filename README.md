#Requisites
- go 1.14+
- postgres
####install go:
- Linux https://dl.google.com/go/go1.14.1.windows-amd64.msi 
- Mac OS https://dl.google.com/go/go1.14.1.darwin-amd64.pkg
- Windows https://dl.google.com/go/go1.14.1.darwin-amd64.pkg

###Install and configure Postgresql:
https://www.thegeekstuff.com/2009/04/linux-postgresql-install-and-configure-from-source/

#Install
Clone this repository:``git clone http://wankar.com:3000/kaypi/kaypi_back_geo.git``

Go to the file:
``
cd kaypi_back_geo/
``

Install go dependencies: 
``
go mod vendor
``

#Run Test
Run this command int the application directory (environment vars are necessary):
``
go test ./
``

#Run Application
##Manually
Export the environment vars to connect with the postgresql database:

	 DB_USER //postgres username
	 DB_PASS //user password
	 DB_NAME //name of the database
	 DB_HOST // host
	 DB_SSL_MODE //ssl mode
	 ENDPOINT // endpoint of the application Ex( /location )
	 SERVER_ADDRESS server address Ex( localhost:8080 )
	 
Run the fallow command to run the application:	 ``go run ./``
##Using the run.sh script
Open the `run.sh` script and change the values of the environment vars and run `sh run.sh`.


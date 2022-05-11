Content of Go-Source-Code-file:

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

type Book struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

func initDb() *sql.DB {

	Db, err := sql.Open("postgres", "host=localhost  port=6000 user=postgres dbname=book password=postgresdb  sslmode=disable")

	checkErr(err)

	return Db
}

func GetAllBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := initDb()
	rows, err := db.Query("SELECT * From books")
	checkErr(err)

	var books []Book

	for rows.Next() {
		book := Book{}
		err = rows.Scan(&book.Id, &book.Title)

		checkErr(err)

		books = append(books, book)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(books)

}

func GetSingleBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	Vars := mux.Vars(r)
	id := Vars["id"]
	var book Book

	db := initDb()

	err := db.QueryRow("SELECT * From books where id = $1", id).Scan(&book.Id, &book.Title)
	checkErr(err)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)

}

func AddBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("Endpoint Hit")

	var book Book

	db := initDb()
	reqBody, err := ioutil.ReadAll(r.Body)
	checkErr(err)
	err = json.Unmarshal(reqBody, &book)
	checkErr(err)

	statement := "insert into books(title) values ($1) returning id"
	stmt, err := db.Prepare(statement)
	checkErr(err)
	defer stmt.Close()
	err = stmt.QueryRow(book.Title).Scan(&book.Id)
	checkErr(err)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("Created")
}

func UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	checkErr(err)

	var book Book

	reqBody, err := ioutil.ReadAll(r.Body)
	checkErr(err)
	err = json.Unmarshal(reqBody, &book)
	checkErr(err)
	db := initDb()
	book.Id = id
	_, err = db.Exec("Update books set title = $2 where id =$1", book.Id, book.Title)
	checkErr(err)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Updated")

}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	id := vars["id"]

	db := initDb()

	_, err := db.Exec("DELETE From books where id = $1", id)

	checkErr(err)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Deleted")

}

func main() {

	mx := mux.NewRouter().StrictSlash(true)

	mx.HandleFunc("/books", GetAllBooks).Methods("GET")
	mx.HandleFunc("/books/{id}", GetSingleBook).Methods("GET")
	mx.HandleFunc("/book", AddBook).Methods("POST")
	mx.HandleFunc("/books/{id}", UpdateBook).Methods("PUT")
	mx.HandleFunc("/books/{id}", deleteBook).Methods("Delete")

	log.Println("Server is Running")
	err := http.ListenAndServe(":8080", mx)
	if err != nil {
		log.Fatalln("Error Starting Http Server:", err)
	}

}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}


Content of Dockerfile:

FROM golang:1.14-alpine3.12 as build


WORKDIR /app

COPY  .  .

RUN go get github.com/gorilla/mux 
RUN go get github.com/lib/pq  

RUN CGO_ENABLED=0 
RUN go build -o  /bin/app
FROM alpine
RUN apk add ca-certificates

COPY --from=build  /bin/app /bin/app

EXPOSE  8080


ENTRYPOINT ["/bin/app"]



Content of Docker-Compose-file:

version: '3.7'
services:
  postgres:
    image: postgres:14.2-alpine
    environment:
        - "POSTGRES_USER:postgres"
        - "POSTGRES_HOST:postgres"
        - "POSTGRES_PORT:5432"
        - "POSTGRES_DB:/var/lib/postgresql/data/book.db"
        - "POSTGRES_PASSWORD=postgresdb"
    ports: 
      - 6000:5432
    restart: always
    volumes:
    - ./data:/var/lib/postgresql/data/

  server:
    image: golang:alpine
    working_dir: /app
    command: go build main.go
    build: 
      context:  .
      dockerfile: Dockerfile
    depends_on:
    - postgres
    environment:
    - "USER:postgres"
    - "HOST:postgres"
    - "PORT:5432"
    - "DBNAME:book"
    - "PASSWORD=postgresdb"
 
    
    ports:
      - 8080:8080
 
volumes:
     data:
   
   
   Output of the Docker-Compose build:
   
   romi@ubuntu:~/Desktop/myapp$ sudo docker-compose up  --build
[sudo] password for romi: 
Building server
Step 1/12 : FROM golang:1.14-alpine3.12 as build
 ---> 7f2c24e45ab4
Step 2/12 : WORKDIR /app
 ---> Using cache
 ---> cc8a3ae605b1
Step 3/12 : COPY  .  .
 ---> d7ab5fb929e6
Step 4/12 : RUN go get github.com/gorilla/mux
 ---> Running in 4ca875ed73ef
go: downloading github.com/gorilla/mux v1.8.0
Removing intermediate container 4ca875ed73ef
 ---> 9ef809ea0031
Step 5/12 : RUN go get github.com/lib/pq
 ---> Running in e3c05769672f
go: downloading github.com/lib/pq v1.10.5
Removing intermediate container e3c05769672f
 ---> fede4915cb07
Step 6/12 : RUN CGO_ENABLED=0
 ---> Running in 4a7a2502cae4
Removing intermediate container 4a7a2502cae4
 ---> 31610c232e9b
Step 7/12 : RUN go build -o  /bin/app
 ---> Running in bf74f7749eea
Removing intermediate container bf74f7749eea
 ---> 4cc9d45aaa1f

Step 8/12 : FROM alpine
 ---> 76c8fb57b6fc
Step 9/12 : RUN apk add ca-certificates
 ---> Using cache
 ---> 024d602d10dd
Step 10/12 : COPY --from=build  /bin/app /bin/app
 ---> Using cache
 ---> eff8b1e93379
Step 11/12 : EXPOSE  8080
 ---> Using cache
 ---> 6f2beba6b309
Step 12/12 : ENTRYPOINT ["/bin/app"]
 ---> Using cache
 ---> 7910a9c881e7

Successfully built 7910a9c881e7
Successfully tagged golang:alpine
Recreating myapp_postgres_1 ... done
Recreating myapp_server_1   ... done
Attaching to myapp_postgres_1, myapp_server_1
postgres_1  | 
postgres_1  | PostgreSQL Database directory appears to contain a database; Skipping initialization
postgres_1  | 
postgres_1  | 2022-05-11 01:43:51.087 UTC [1] LOG:  starting PostgreSQL 14.2 on x86_64-pc-linux-musl, compiled by gcc (Alpine 10.3.1_git20211027) 10.3.1 20211027, 64-bit
postgres_1  | 2022-05-11 01:43:51.087 UTC [1] LOG:  listening on IPv4 address "0.0.0.0", port 5432
postgres_1  | 2022-05-11 01:43:51.088 UTC [1] LOG:  listening on IPv6 address "::", port 5432
postgres_1  | 2022-05-11 01:43:51.093 UTC [1] LOG:  listening on Unix socket "/var/run/postgresql/.s.PGSQL.5432"
postgres_1  | 2022-05-11 01:43:51.101 UTC [21] LOG:  database system was shut down at 2022-05-11 01:43:48 UTC
postgres_1  | 2022-05-11 01:43:51.114 UTC [1] LOG:  database system is ready to accept connections
server_1    | 2022/05/11 01:43:53 Server is Running


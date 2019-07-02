# dito-desafio

Comando para execução: ```docker-compose up```

Endpoints:

* POST http://localhost:8080/event 

    {
        "event": "buy",
        "timestamp": "2016-09-22T13:57:31.2311892-04:00"
    }
* GET http://localhost:8080/event?event=bu
* GET  http://localhost:8080/groupEvents

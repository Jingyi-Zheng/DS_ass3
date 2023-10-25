# DS_Ass3

This is the ASSIGNMENT 3 for the Distributed Systems course at ITU (Copenhagen).

- How to run the server?
   - Execute "go run server.go" in the server folder. It defaults to tcp communication and uses port 50051 (of course, you can always change it in server.go)

- How to run the clients？
-   Execute "go run client.go -cPort 8080 -sPort 50051" in client folder. The parameter after cPort is the port number used by the client that is turned on, and the parameter after sPort is the port that server used which client goes to connect to.

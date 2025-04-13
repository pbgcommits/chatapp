# chatapp
Simple chat app to practise concurrency/networking concepts in go.

## Features
- Easy to use - once the server is started, join from the shell using ```nc```
- Message individual users, or all logged in users
- Authentication - join with a username and password
- Multiple users can log in and interact with each other
- Multiple simultaneous sessions for one user - log in as the same user from multiple shells

## How to run
1. Download the repository to your local machine.
2. Open a shell in the repository and run 
```
go run .
```
to start the server.

3. Open a new shell from any repo and run
```
nc localhost 9018
```
 To connect to the server as a new client.

4. Set a username and password, or login as an existing user (note - these are currently
 cleared when the server restarts)
5. 
- Send a message to all online users by default. 
- Send a message to specific users by typing -u user1:user2:... (message)
- Get help by typing -h

## TODOS
In the future this will have:
- A database to store users who have joined previously
- A simple GUI
# pingedchat
A self-hosted chat client built in golang

#### General Overview
Conceptually, pingedchat uses these technologies:
- golang server backend
-- sockjs
-- storage database, I used postgres
-- quick-access database, I used aerospike
-- messaging connect, I used gnatsd

- web client requirements
-- sockjs
-- local database, I used lokidb

#### Architectural Overview


  
|--------------------------|  
|-  database_functions.go -| (Run() will call a function defined in database_functions.go based on the command)  
|--------------------------|  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;^  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|  
|----------------|  
|-  database.go -| (Run(), this is the main thread for validated user)  
|----------------|  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;^  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|   
|------------|&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|------------------|  
|-  conn.go -| (sockHandler(), validates the user) ->    |- database.go -| (Connect(), opens database connections)  
|------------|&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|------------------|  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;^  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|  
|------------|  
|-  main.go -| (sockjsServerLoop)  
|------------|  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;^  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|  
|--------------|  
|- Web client -|  
|--------------|  


I'm a visual learner, so I "drew" that out to try and demonstrate what is happening.
A user loads the static webpage (no generated pages here!), and the sockjs client connects to our sockjsServerLoop() function.  This in turn calls SockHandler(), which reads in whether the user wishes to register (CreateUser) or login (ValidateUser).  Once the user is either validated successfully or registered successfully, we enter the main Run() loop in database.go .  This loop will run as long as the user is connected, and is contantly listening for commands.  The large "functs" map variable stores the possible commands to function mapping.  From there, the function is run, and if the function wishes to return a string to the user, it returns that over the sockjs channel.


#### Databases
I used two databases, postgres and aerospike.  Postgres works well for storing messages, which were done with one table per conversation thread.  I didn't want to store every message in one large database, and didn't want to save the message more than once.  Saving the messages in their own table meant whenever the user logs in, I do a "SELECT * FROM CONVERSATION" for each conversation in the user.  This also allows me to update each conversation individually when a user refreshes the page, since the user keeps track of the last modified time (m_time).
I chose aerospike over redis because of their speed boasts, but really any NoSQL database would work, similar to how any SQL database could replace Postgres for our use cases.  The user's data is stored in the aerospike database, and the fields can be found in structs.go , UserStruct{}.


#### Scheduled messages
There is support for scheduled messages.  This is simply executed by a goroutine checking a postgres database at a predetermined time interval (2 minutes), and removing any messages that were scheduled to be sent in the past and removing them.  Since we always remove messages in the past, we can just keep removing them when we query the database.  The messages are then added to the correct Postgres conversation table in SendMessage() .


#### Other notes
- add_deps.sh will be handy if you try to run this on your own server.  Takes away the "go get"'s.
- aerospike.conf is a sample configuration for aerospike that I used for this project.  The last configuration was for a linode server setup.
- There should be "FILL_ME" wherever an API key is needed.
- The Protect() function takes in a function as a parameter, runs the function, and catches any panic it throws.  It then logs the panic and restarts the function.
- I used [ffjson](https://github.com/pquerna/ffjson) for most structs in this project.  Since I kept all structs in their own files (structs.go , mailStructs.go) the corresponding *_ffjson.go files were automatically generated by the ffjson utility.
- I had used Mandrill as an inbound email webhook service, but they changed their terms of usage after I stopped working on this.  Not sure how usable they are now for that purpose, but there are alternatives out there.
- All comments were written while actively working on the code.
- The code is set up to run over HTTPS/WSS, but I did not include any server certificates.  They can be self-generated for testing, purchased from a CA, or you can disable the TLS support by modifying the HTTP server code.  When serving over HTTP, WS is automatically used; HTTPS results in automatic WSS.


#### Questions
I'm sure I missed something, so either email me with a question or make an issue and I'll do my best to respond :)

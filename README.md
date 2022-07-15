# httpsipmon

Simple daemon to check SIP status via HTTP

Runs a http server which will send a SIP OPTIONS message each time the server is called.
The http server returns the status code from the remote SIP server
Can be used to check a remote sip servers status with http.
```
Monitoring Agent     httpsipmon          sip server 
       |                  |                    | 
       |  -- http GET --> |                    |
       |                  | -- sip OPTIONS --> |
       |                  |     <-- 200 OK --  |
       |  <-- 200 OK --   |                    |
```

## Prereq 
You will need go installed 

## Install 
```
go install github.com/footfish/httpsipmon@latest
```
## Run it
```
# assuming go install put it in go/bin
cd go/bin 
# connect to sip service sip.linphone.org op port 5060 
./httpsipmon sip.linphone.org:5060 &  

```
## Check it
```
#note that you will likely get 403 forbidden/404 not found unless the SIP server has been configured to accept your call 
curl -v localhost:8080
```

Uses https://github.com/jart/gosip
# Gockets
**Gockets** is daemon written in Golang to give languages, like PHP a middleware for REST-oriented communication with Websockets.

## Build
1. Clone repository
2. Change current directory to repository folder
3. Execute `go build main.go`

## Running

Execute binary created after building. Pass `-h` to get help about arguments  
    `-config string`  
	Path to configuration file. Default `config.yml`  
## Usage
For convenience [Postman collection](https://documenter.getpostman.com/view/4929623/S1LpbBvp) was published with latest API reference.  

By default daemon is listening to **8844** port. Change it in `main` function if needed.

`POST` `/channel/prepare` - prepares channel and returns JSON object which contains `publisher_key` and `subscriber_key`.  May contain JSON object in body with hook url specified. Example:
```json
{
	"subscriber_message_hook_url" : "http://localhost/log.php",
	"tag" : "someApplication|priority:5|thirdTag"
}
```
`GET` `/channel/show` - list of all channels awaiting publishing or subscription. May be filtered by tag with `search` parameter by passing URLEncoded regex  
`GET` `/channel/show/{publisher_key}` - get specific channel data.    
`GET` `/channel/subscribe/{subscriber_key}` - creates a Websocket connection with channel referenced by `subscriber_key`.  
`PATCH` `/channel/edit/{publisher_key}` - edit channel by passing body of same structure as for `POST` `/channel/prepare` request.  
`POST` `/channel/publish/{publisher_key}` - pushes data to a Websocket connection passed in `body` of request.  
`DELETE` `/channel/publish/{publisher_key}` - closes all Websocket connection to channel specified and deletes channel itself.

## Use case

Create a channel and share a subscriber key to any Front-End application you want to communicate via Websocket. After doing so, with any kind of mechanisms, like Events in Laravel Framework hook pushing data into Websocket.

### TODO

* ~~Implementation of graceful close of Websocket connection~~
* ~~Implementation of 2-way communication with URL-Hook~~
* Refactoring of already created methods
* Better error handling
* Callback hook refactor
* Tests
* Binary data support





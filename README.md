## **GO-HORSE** : DOCKER DAEMON PROXY/FILTER

>The software in the middle the communication between docker client and daemon, allowing you to intercept all commands and, by example, do access control or add tags in a container during its creation, change its name, alter network definition, redifine volumes, rewrite the whole body if you want, and so on. Take the control. Do what you need.

#### Table of contents

1. [ How it works ](#how_it_works)
2. [ Running ](#running)
3. [ Filtering requests using JavaScript ](#js_filter)
4. [ Filtering requests using Go ](#go_filter)

<br/>

<a name="how_it_works"/>
### How it works

Docker (http) commands sent from the client to the deamon are intercepted by creating filters in go-horse. This filters can be implemented either in JavaScript or Golang. You should inform a *path pattern* to match a command url (check [docker api docs](https://docs.docker.com/engine/api/v1.39/) or see go-horse logs to map what urls are requested by docker client commands), a *invoke* property telling if you want the filter to run at the Request time, before the request hit the daemon, or on Response time, after the daemon has processed the request. Once your filter gets a request, you have all the means to implement the rules your business needs. Rewrite a url to the docker daemon? Check the user identity in another system? Send a http request and break the filter chain based on the response? Add metadata to a container? Change container properties? Compute specific metrics?  Blacklist some commands? Ok, can do. This and many more.

<a name="running"/>
### Running

```docker-compose
version: '3.7'
services:
  proxy:
    image: labbs/go-horse
    network_mode: bridge
    ports: 
      - 8080:8080
    environment: 
      - DOCKER_HOST=/var/run/docker.sock
      - DOCKER_SOCK=unix:///var/run/docker.sock
      - LOG_LEVEL=debug
      - PRETTY_LOG=true
      - PORT=:8080
      - JS_FILTERS_PATH=/app/go-horse/filters
      - GO_PLUGINS_PATH=/app/go-horse/plugins
    volumes: 
      - /var/run/docker.sock:/var/run/docker.sock
      - /home/bruno/go-horse:/app/go-horse
```
Set the environment variable `DOCKER_HOST` to `tcp://localhost:8080` or test a single command adding -H attribute to a docker command : `docker -H=localhost:8080 ps -a` and watch the go-horse container logs

<br/>
<a name="js_filter"/>
### Filtering requests using JavaScript
According to the environment variable `JS_FILTERS_PATH`, you have to place your JavaScript filters there to get them loaded in the go-horse filter chain. These file's name have to obey the following pattern :

- `000.request.test.js` => {order}.{invoke}.{name}.{extension}

| Property  | Values | 000.request.test.js | Description|
| ------------- | ------------- |------------| ------------|
| Order  | [0-9]{1,3} | `000` | Filter execution order is sorted by this property and should be unique.| 
| Invoke  | `request` or `response` | `request` |  Filter will be invoked before(Request) or after(Response) the command was sent to daemon|
| Name | .* | `test` | A name for your filter |
| Extension | `js` | `js` |Fixed - mandatory |

Create a file with the convention above, place it in the right directory - remember the `JS_FILTER_PATH` and paste the following code : 

```
{
	"pathPattern": ".*",
	"function" : function(ctx, plugins) {
		console.log(">>> hello, go-horse");
		return {status: 200, next: true, body: ctx.body, operation : ctx.operation.READ};
	}
}
```

Before executing a docker command, check the go-horse logs again. 

Did you see it? Yeah! Live reloading for JS filters. Nice, uh? No? We did this trying to help you during filters development and also don't let the SysAdmins down when everything else is. If that bothers you, you can build a docker image `FROM labbs/go-horse` including the filters in its file system. And there you go, immutable happiness all the way.

Now run a docker command like `docker image ls`. Watch the logs again. You should see something like this :

```
4:17PM DBG Receiving request request="[4] ::1 ▶ GET:/_ping"
4:17PM DBG Request the mainHandler: /_ping
>>> hello, go-horse
4:17PM DBG Executing request for URL : /_ping ...
4:17PM DBG Response the mainHandler:/_ping
4:17PM DBG Receiving request request="[4] ::1 ▶ GET:/v1.39/images/json"
4:17PM DBG Request the mainHandler: /v1.39/images/json
>>> hello, go-horse
4:17PM DBG Executing request for URL : /v1.39/images/json ...
4:17PM DBG Response the mainHandler:/v1.39/images/json
```

We intercepted every request to docker daemon as configured by the property **pathPatter** in the filter definition file with the regex `.*`. Even though this is being a JavaScript file, that property's value will be used in the Go context (`regexp.Regexp`) to filter the URLs, so don't use JS regexes, they won't work in go-horse. Sorry. Test your patterns in sites like https://regex101.com/ with the golang ?flavor? selected.

Now look at the `function` function - Yes, naming things aren't one of our strengths. You will see more of this as you continue reading and get more involved with go-horse. Let us explain how this `function` function works: (there's a whole functionality going on)

That function called as 'function' receives 2 arguments. The first one, `ctx` has data and functions provided by go-horse, it is related to the 'client and daemon communication' and filter chain. The second one, the `plugins` argument, will contain data and functions provided by you. It's a way to extend the filter's context, if you need it. Letting you inject all things we forgot to include. We explain that better. Later. Now, more about the `ctx` variable and their properties :

##### Filter function arguments

| ctx.`Property`  | Type       | Description| Parameters | Return | 
| --------- | ---------- |------------|------------|------------|
|ctx.**url**|string|original url called by docker client|-|-|
|ctx.**body**|object|body of the request from the client or the body's response from the daemon. Depending on the `invoke` field in the filter's file definition name|-|-|
|ctx.**operation**|object|a helper object to use in the return of the filter function `function`, telling if the body should be **overriden** :`operation.WRITE` or **not** : `operation.READ`|-|-|
|ctx.**method**|string|http method of the request from the client|-|-|
|ctx.**getVar**|function|get the value of this variable with scope limited by the request lifetime and shared between all filters| -  [string] var name|- [string] var value
|ctx.**setVar**|function|set the variable with the provided value and make that avaliable to the next filters in the chain until the end of the request| - [string] name <br/> -  [string] value |-|
|ctx.**listVar**|function|list all variable names within this request's scope|-|[string array] names
|ctx.**headers**|object|original headers sended by docker client|-| [map string string]
|ctx.**request**|function|as we saw earlier, another bad name! They have spread all over - easy pull requests, just to mention... that function executes a http request | - [string] http method <br/> - [string] url <br/> - [string] body <br/> - [object] headers <br/>| [object] -> [body : object], [status : int], [headers : object] |

After process the request, the filter needs to return a object like this :

##### Filter function return

`{status: 200, next: true, body: ctx.body, operation : ctx.operation.READ}`

| Property  | Values | Example | Description|
| ------------- | ------------- |------------| ------------|
| status  | int | `200` | In case of error, to overwrite original status.| 
| next  | boolean | `true` | This property tells go-horse to stop the filter chain and don't run other filters after this. |
| body | object | `ctx.body` | Only useful when you need to substitute the original |
| operation | `ctx.operation.READ` or `ctx.operation.WRITE` | `ctx.operation.READ` | READ : does nothing, next filter receive the same body as you did; WRITE : pass the body property you modified to the next filters or send to the docker client if your filter is the last in the chain |

##### Rewriting URLs sended to the daemon

There's a special variable stored in the request scope that should be changed if you need to rewrite the URL used to daemon's requests : `targetEndpoint`. The way to alter it value is to call the setVar function in the ctx object, argument of the filter function : `ctx.setVar('targetEndpoint', '/v1.39/newEndpoint')`.
This was useful when we needed to pass a token in the DOCKER_HOST environment variable to identify the user. Ther token was extracted, verified against other system and the original URL was restored (if user was authorized), because the daemon doesn't like tokens.

##### Environment variables in JS filters

All env vars are avaliable in javascript filters scope. You can list them by calling `ctx.listVars` method. They are have an 'ENV_' prefix.

<br/>
<a name="go_filter"/>
### Filtering requests using Go

Besides Javascript, you can also create your filters using GoLang. If you doesn't like JS, if you don't want to be constraint by JS context limitation, if you care about performance or ... ?? then use Go Filters. 

### Extending Javascript filter context with Go Plugins

### JS versus GO - information to help your choice
--- benchmark



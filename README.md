## **GO-HORSE** : DOCKER DAEMON PROXY/FILTER

>The software in the middle the communication between docker's client and daemon, allowing you to intercept all commands and, by example, do access control or add tags in a container during its creation, change its name, alter network definition, redifine volumes, rewrite the whole command's request body if you want, and so on. Take the control. Do what you need.


#### Table of contents

1. [ How it works ](#how_it_works)<br/>
2. [ Running ](#running)<br/>
   2.1. [ Environment variables ](#envvars)<br/>
3. [ Filtering requests using JavaScript ](#js_filter)<br/>
  3.1. [ Filter function arguments ](#js_filter_func_args)<br/>
  3.2. [ Filter function return ](#js_filter_func_ret)<br/>
  &emsp;3.2.1. [Tricky return combinations](#js_tricky)<br/>
  3.3. [ Rewriting URLs sent to the daemon ](#js_rewrite)<br/>
  3.4. [ Environment variables in JS filters ](#js_env_vars)<br/>
4. [ Filtering requests using Go ](#go_filter)<br/>
  4.1. [ Go filter interface ](#go_interface)<br/>
  4.2. [ Sample Go filter ](#go_sample)<br/>
  4.3. [ Compiling and running a golang filer ](#go_compile)<br/>
  4.4. [ Another go filter sample ](#go_sample_2)
5. [ Extending Javascript filter context with Go Plugins ](#go_plugin)
6. [ JS versus GO - information to help your choice ](#benchmark)

<br/>
<a name="how_it_works"/>

### 1. How it works

Docker (http) commands sent from the client to the daemon are intercepted by creating filters in go-horse. This filters can be implemented either in JavaScript or Golang. You should inform a *path pattern* to match a command url (check [docker api docs](https://docs.docker.com/engine/api/v1.39/) or see go-horse logs to map what urls are requested by docker client commands), a *invoke* property telling if you want the filter to run at the Request time, before the request hit the daemon, or on Response time, after the daemon has processed the request. Once your filter gets a request, you have all the means to implement the rules your business needs. Rewrite a url to the docker daemon? Check the user identity in another system? Send a http request and break the filter chain based on the response? Add metadata to a container? Change container properties? Compute specific metrics?  Blacklist some commands? Ok, can do. This and many more.

<br/>
<a name="running"/>

### 2. Running

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

<a name="envvars"/>

#### 2.1. Environment variables

Besides the self explanatory variables, there are : 

| Env Var         | Type    | Description                                        |
| --------------- |  -------| ---------------------------------------------------|
| PRETTY_LOG      | boolean | logs by default are ~~ugly~~ in json format        |
| JS_FILTER_PATH  | path    | where, in the images file system, are the js filter|
| GO_PLUGINS_PATH | path    | where, in the images file system, are the go filter and the go plugins|

<br/>
<a name="js_filter"/>

### 3. Filtering requests using JavaScript
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

<a name="js_filter_func_args"/>

##### 3.1. Filter function arguments

| ctx.`Property`  | Type       | Description| Parameters | Return | 
| --------- | ---------- |------------|------------|------------|
|ctx.**url**|string|original url called by docker client|-|-|
|ctx.**body**|object|body of the request from the client or the body's response from the daemon. Depending on the `invoke` field in the filter's file definition name|-|-|
|ctx.**operation**|object|a helper object to use in the return of the filter function `function`, telling if the body should be **overridden** :`operation.WRITE` or **not** : `operation.READ`|-|-|
|ctx.**method**|string|http method of the request from the client|-|-|
|ctx.**values**|object|a object with functions to share data between all filters by request lifetime|-|-|
|ctx.values.**get**|function|get the value of this variable with scope limited by the request lifetime and shared between all filters| -  [string] var name|- [string] var value
|ctx.values.**set**|function|set the variable with the provided value and make that available to the next filters in the chain until the end of the request| - [string] name <br/> -  [string] value |-|
|ctx.values.**list**|function|list all variable names within this request's scope|-|[string array] names
|ctx.**urlParams**|object|a object with functions to manipulate request query parameters|-|-|
|ctx.urlParams.**add**|function|adds the value to key. It appends to any existing values associated with key.| -  [string] var key|-|
|ctx.urlParams.**get**|function|gets the value associated with the given key. If there are no values associated with the key, get returns the empty string| -  [string] var key|- [string] var value|
|ctx.urlParams.**set**|function|sets the key to value. It replaces any existing value| - [string] key |-|
|ctx.urlParams.**del**|function|deletes the values associated with key| - [string] key |-|
|ctx.urlParams.**list**|function|parses query parameters and returns an object with corresponding key-value|-|[object] values
|ctx.**responseStatusCode**|string|original status code from daemon http response|-| [string] status code
|ctx.**headers**|object|original headers sent by docker client|-| [map string string]
|ctx.**request**|function|as we saw earlier, another bad name! They have spread all over - easy pull requests, just to mention... that function executes a http request | - [string] http method <br/> - [string] url <br/> - [string] body <br/> - [object] headers <br/>| [object] -> [body : object], [status : int], [headers : object] |

After process the request, the filter needs to return a object like this :

<a name="js_filter_func_ret"/>

##### 3.2. Filter function return

`{status: 200, next: true, body: ctx.body, operation : ctx.operation.READ}`

| Property  | Type | Example | Description|
| ------------- | ------------- |------------| ------------|
| status  | int | `200` | In case of error, to overwrite original status.| 
| next  | boolean | `true` | This property tells go-horse to stop the filter chain and don't run other filters after this. |
| body | object | `ctx.body` | Only useful when you need to substitute the original |
| operation | `ctx.operation.READ` or `ctx.operation.WRITE` | `ctx.operation.READ` | READ : does nothing, next filter receive the same body as you did; WRITE : pass the body property you modified to the next filters or send to the docker client if your filter is the last in the chain |

<a name="js_tricky"/>

##### 3.2.1. Tricky return combinations

**`return { error: "something bad happen" }`**

go-horse will assume the following default values : status = 0; next = false; body = "", operation = READ(0) and the docker client will print in the terminal : "something bad happen". Filter chain will stop.

**`return { body : "something bad happen", status : 500 }`**

go-horse will assume the following default values : next = false; operation = READ(0) and the docker client will print in the terminal : "something bad happen". Filter chain will stop.

**`return { next: true, body : "something bad happen", status : 500 }`**

go-horse will assume the following default values : operation = READ(0). Filter chain won't stop and if another filter don't override the body and error status, docker client will print in the terminal : "something bad happen"

**`return { next: true, body : [{ container : ... }], status : 200, operation : ctx.operation.READ }`**

go-horse will ignore the body you returned because of the value `ctx.operation.READ` is set in response's *operation* field. Next filter will receive the same body you has received.

<a name="js_rewrite"/>

##### 3.3. Rewriting URLs sent to the daemon

There's a special variable stored in the request scope that should be changed if you need to rewrite the URL used to daemon's requests : `path`. The way to alter it value is to call the setVar function in the ctx object, argument of the filter function : `ctx.values.set('path', '/v1.39/newEndpoint')`.
This was useful when we needed to pass a token in the DOCKER_HOST environment variable to identify the user. The token was extracted, verified against other system and the original URL was restored (if user was authorized), because the daemon doesn't like tokens.

<a name="js_env_vars"/>

##### 3.4. Environment variables in JS filters

All env vars are available in javascript filters scope. You can list them by calling `ctx.values.list()` method. They are have an 'ENV_' prefix.

##### 3.5. Passing a token in DOCKER_HOST url

**WARNING** this may change in a near future. We are not comfortable with this solution as is.

The routes are duplicated, one version with a `/token/{token}` prefixed and another version without the token prefix.

In our case, this was used in conjunction with a CLI that logs the user in and changes the DOCKER_HOST env var in the user machine to our go-horse address with the token added in its path. The first filter - with order equals 0, extract the token, validates the user and rewrite the url calling `ctx.values.set('path', '<tokenless_url>')`.

Another possible solution, and more elegant - i think, is to insert a token as a header in all docker CLI commands requests. This can be achieved by editing /~/.docker/config.json file, inserting the property `"HttpHeaders": { "token": "?" },`. The request to docker daemon will carry the token in its headers and a filter can read and validate it - sending a request to an identity manager? Maybe.

<br/>
<a name="go_filter"/>

### 4. Filtering requests using Go

Besides Javascript, you can also create your filters using GoLang. If you don't like JS, if you don't want to be constrained by JS context limitation, ~~if you care about performance~~ (check out our surprisingly [ benchmark results ](#benchmark)) or ... ?? then, use Go Filters. It is up to you.

<a name="go_interface"/>

##### 4.1. Go filter interface

```go
type Filter interface {
	Config() (Name string, Order int, PathPattern string, Invoke int)
	Exec(ctx iris.Context, requestBody string) (Next bool, Body string, Status int, Operation int, Err error)
}
```
Your go filters have to implement those functions.

The `Config` method tells go-horse information about yout filter. Same rules as [ JavaScript filters ](#js_filter).

Operation => 0 = Read<br/>Operation => 1 = Write

The `Exec` method, runs when a request hits go-horse and his url match the `Config.PathPattern` attribute.

Invoke => 0 = Response<br/>Invoke => 1 = Request

<a name="go_sample"/>


##### 4.2. Sample GO filter

Create a go file named *sample_filter.go* .

``` go
package main

import (
	"fmt"
	"github.com/kataras/iris"
)

func main() {}

// PluginModel PluginModel
type PluginModel struct {
	Next      bool
	Body      string
	Status    int
	Operation int
}

// Filter Filter
type Filter interface {
	Config() (Name string, Order int, PathPattern string, Invoke int)
	Exec(ctx iris.Context, requestBody string) (Next bool, Body string, Status int, Operation int, Err error)
}

// Exec Exec
func (filter PluginModel) Exec(ctx iris.Context, requestBody string) (Next bool, Body string, Status int, Operation int, Err error) {
	fmt.Println(">>> body response from docker daemon >>> ", requestBody)
	return true, "newBody: i'm sure almost everyBody needs one", 500, 1, nil
}

// Config Config
func (filter PluginModel) Config() (Name string, Order int, PathPattern string, Invoke int) {
	return "GO_FILTER", 0, ".*", 0
}

// Plugin exported as symbol
var Plugin PluginModel

```

<a name="go_compile"/>

##### 4.3. Compiling and running a golang filer

Save the file above and run the following command in terminal to compile it :

`go build -buildmode=plugin -a -installsuffix cgo -o sample-filter.so sample_filter.go`

Copy the `sample-filter.so` to `GO_PLUGINS_PATH` directory. Restart go-horse, run `docker ps -a` command. You should see something like this in the logs : 

``` terminal
5:06PM INF Receiving request request="[1] ::1 ▶ GET:/_ping"
5:06PM DBG Running REQUEST filters for url : /_ping
5:06PM DBG Executing request for URL : /_ping ...
5:06PM DBG Running RESPONSE filters for url : /_ping
5:06PM DBG executing filter ... Filter matched="[1] ::1 ▶ GET:/_ping" filter_config="model.FilterConfig{Name:\"GO_FILTER\", Order:0, PathPattern:\".*\", Invoke:0, Function:\"\", regex:(*regexp.Regexp)(nil)}"
>>> request body >>>  OK
5:06PM DBG filter execution end Filter output="model.FilterReturn{Next:true, Body:\"newBody: i'm sure almost everyBody needs one\", Status:500, Operation:1, Err:error(nil)}" filter_config="model.FilterReturn{Next:true, Body:\"newBody: i'm sure almost everyBody needs one\", Status:500, Operation:1, Err:error(nil)}"
5:06PM DBG Body rewrite for filter : GO_FILTER
5:06PM INF Receiving request request="[2] ::1 ▶ GET:/v1.39/containers/json?all=1"
5:06PM DBG Running REQUEST filters for url : /v1.39/containers/json
5:06PM DBG Executing request for URL : /v1.39/containers/json?all=1 ...
5:06PM DBG Running RESPONSE filters for url : /v1.39/containers/json
5:06PM DBG executing filter ... Filter matched="[2] ::1 ▶ GET:/v1.39/containers/json?all=1" filter_config="model.FilterConfig{Name:\"GO_FILTER\", Order:0, PathPattern:\".*\", Invoke:0, Function:\"\", regex:(*regexp.Regexp)(nil)}"
>>> request body >>>  [{"Id":"92e54dd9478cc8dc5f36173c5f4ee3de3875c20ec1cab55a6c8e6f9597823cd1","Names":["/go-horse_proxy_1_cee5d01ac3ca"],"Image":"sandman_proxy","ImageID":"sha256:8caa5ad78baea70ddb4bfc7830b3a61f604258e8a23d2cd4857f806cd29cbb40","Command":"/main","Created":1546035689,"Ports":[],"Labels":{"com.docker.compose.config-hash":"6559e4771dd522f04288450e8c42fd98499cebfa1169f7dafcdc0e8eeff04418","com.docker.compose.container-number":"1","com.docker.compose.oneoff":"False","com.docker.compose.project":"go-horse","com.docker.compose.service":"proxy","com.docker.compose.slug":"cee5d01ac3caaaf9425f6e240691b3dbb88c2bfc44e0d720c2dd061ac903fbc","com.docker.compose.version":"1.23.1"},"State":"created","Status":"Created","HostConfig":{"NetworkMode":"bridge"},"NetworkSettings":{"Networks":{"bridge":{"IPAMConfig":null,"Links":null,"Aliases":null,"NetworkID":"7c5e3129b1a5922b2d73dd88a02c154f587cc283a85217e6ce99ab4b8d8ac020","EndpointID":"e1dcdcbc8fb3e0c3fdab76e62415f457cbfd2978c8c7c336a915e4fab706f5b1","Gateway":"","IPAddress":"192.168.1.2","IPPrefixLen":24,"IPv6Gateway":"","GlobalIPv6Address":"","GlobalIPv6PrefixLen":0,"MacAddress":"02:42:c0:a8:01:02","DriverOpts":null}}},"Mounts":[{"Type":"bind","Source":"/home/bruno/sadman-acl-proxy","Destination":"/app/sadman-acl-proxy","Mode":"rw","RW":true,"Propagation":"rprivate"},{"Type":"bind","Source":"/var/run/docker.sock","Destination":"/var/run/docker.sock","Mode":"rw","RW":true,"Propagation":"rprivate"}]},{"Id":"a638bca1d06a13dfca0460c07ccad05dff83f9318836018e6af5b1f715b5308e","Names":["/teste"],"Image":"redis","ImageID":"sha256:415381a6cb813ef0972eff8edac32069637b4546349d9ffdb8e4f641f55edcdd","Command":"docker-entrypoint.sh redis-server","Created":1546033927,"Ports":[{"IP":"0.0.0.0","PrivatePort":6379,"PublicPort":6379,"Type":"tcp"}],"Labels":{},"State":"exited","Status":"Exited (255) 2 days ago","HostConfig":{"NetworkMode":"default"},"NetworkSettings":{"Networks":{"bridge":{"IPAMConfig":null,"Links":null,"Aliases":null,"NetworkID":"7c5e3129b1a5922b2d73dd88a02c154f587cc283a85217e6ce99ab4b8d8ac020","EndpointID":"bec00b6890735756b60cd3512bb521c71530b7c66ea975d3972deecd11772a1f","Gateway":"192.168.1.5","IPAddress":"192.168.1.1","IPPrefixLen":24,"IPv6Gateway":"","GlobalIPv6Address":"","GlobalIPv6PrefixLen":0,"MacAddress":"02:42:c0:a8:01:01","DriverOpts":null}}},"Mounts":[{"Type":"volume","Name":"591ec570c87431ed7cd7292d0551d386b800456cbca721e36b304b38ca625649","Source":"","Destination":"/data","Driver":"local","Mode":"","RW":true,"Propagation":""}]}]

5:06PM DBG filter execution end Filter output="model.FilterReturn{Next:true, Body:\"newBody: i'm sure almost everyBody needs one\", Status:500, Operation:1, Err:error(nil)}" filter_config="model.FilterReturn{Next:true, Body:\"newBody: i'm sure almost everyBody needs one\", Status:500, Operation:1, Err:error(nil)}"
5:06PM DBG Body rewrite for filter : GO_FILTER
```

And docker client should print this : 

``` terminal
[bruno@labbsr0x go-horse]$ docker ps -a
Error response from daemon: newBody: i'm sure almost everyBody needs one
```

Cool? Let's create another one, this time we will not return an error to docker client.

<a name="go_sample_2"/>

##### 4.3. Another go filter sample

Now we are gonna reverse the container's name and add a label during its creation.

``` go
package main

import (
	"strings"

	"github.com/kataras/iris"
	"github.com/tidwall/sjson"
)

// PluginModel PluginModel
type PluginModel struct {
	Next      bool
	Body      string
	Status    int
	Operation int
}

// Filter Filter
type Filter interface {
	Config() (Name string, Order int, PathPattern string, Invoke int)
	Exec(ctx iris.Context, requestBody string) (Next bool, Body string, Status int, Operation int, Err error)
}

// Exec Exec
func (filter PluginModel) Exec(ctx iris.Context, requestBody string) (Next bool, Body string, Status int, Operation int, Err error) {
	containerName := strings.Split(ctx.Values().GetString("path"), "?name=")
	containerNameReversed := Reverse(containerName[1])
	ctx.Values().Set("path", containerName[0]+"?name="+containerNameReversed)
	value, _ := sjson.Set(requestBody, "Labels", map[string]interface{}{"pass-through": "Go-Horse"})
	return true, value, 200, 1, nil
}

// Config Config
func (filter PluginModel) Config() (Name string, Order int, PathPattern string, Invoke int) {
	return "GO_FILTER_CONTAINER_CREATE_ADD_LABEL", 0, "/containers/create", 1
}

// Plugin exported as symbol
var Plugin PluginModel

// Reverse Reverse
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
```
Compile as you did in the previous filter and place it in the right dir. Restart go-horse again.

Run a `docker run -d --name sample_container redis`

Run a `docker ps`

```
CONTAINER ID        IMAGE               COMMAND                  CREATED             STATUS              PORTS                    NAMES
d6ad06221d8d        redis               "docker-entrypoint.s…"   3 seconds ago       Up 2 seconds        6379/tcp                 reniatnoc_elpmas
```

See the containers name we just created? sample_container => reniatnoc_elpmas

Run a `docker inspect reniatnoc_elpmas | grep -i -C 5 'Labels'`

``` terminal
            "WorkingDir": "/data",
            "Entrypoint": [
                "docker-entrypoint.sh"
            ],
            "OnBuild": null,
            "Labels": {
>>>>>>>>>>>>  "pass-through": "Go-Horse"
            }
        },
        "NetworkSettings": {
            "Bridge": "",
```



<br/>
<a name="go_plugin"/>

### 5. Extending Javascript filter context with Go Plugins

If you need something in JS filter context that is not there, you can create a go plugin to inject anything you need in JS filter through the `plugin` argument in the function `function`.

Go plugin :

``` go
package main

import (
	"fmt"
	"time"

	"github.com/kataras/iris"
	"github.com/robertkrimen/otto"
)

func main() {}

// PluginModel ler-lero
type PluginModel struct{}

// Set lero-lero
func (js PluginModel) Set(ctx iris.Context, call otto.FunctionCall) otto.Value {

	startingTime := time.Now().UTC()

	ready := make(chan bool)
	defer close(ready)

	go func() {
		time.Sleep(2 * time.Second)
		ready <- true
	}()

	obj, err := call.Otto.Object("({})")

	select {
	case msg := <-ready:
		fmt.Println(">>>> GO JS PLUGIN >>>> ready : ", msg)
		obj.Set("ready", msg)
		obj.Set("timeSpentUntilCallThisFunctionSincePluginWasInjected", func() int64 {
			endingTime := time.Now().UTC()
			duration := endingTime.Sub(startingTime)
			return int64(duration)
		})
	}

	if err != nil {
		fmt.Println("erro ao cria o objecto de rectorno da função js do plugin  ", js.Name(), " : ", err)
	}
	return obj.Value()
}

// Name lero-lero
func (js PluginModel) Name() string {
	return "sample"
}

// Plugin exported as symbol named "Greeter"
var Plugin PluginModel
```
And the Javascript filter that use the plugin : 

```javascript
{
	"pathPattern": ".*",
	"function" : function(ctx, plugins){
		console.log(">>>>> plugins.sample().ready : ", plugins.sample().ready)
		console.log(">>>>> plugins.sample().timeSpentUntilCallThisFunctionSincePluginWasInjected() : ", 
			plugins.sample().timeSpentUntilCallThisFunctionSincePluginWasInjected())
		return {status: 200, next: true, body: ctx.body, operation : ctx.operation.READ};
	}
}
```
Now compile the plugin and place the .so file and the js filter in the right folder. Run a `docker ps` command and watch the logs : 

``` terminal
9:24PM DBG executing filter ... Filter matched="[6] ::1 ▶ GET:/v1.39/containers/json?all=1" filter_config="model.FilterConfig{Name:\"plugin_sample\", Order:0, PathPattern:\".*\", Invoke:0, Function:\"function(ctx, plugins){\\n\\t\\tconsole.log(\\\">>>>> plugins.sample().ready : \\\", plugins.sample().ready)\\n\\t\\tconsole.log(\\\">>>>> plugins.sample().timeSpentUntilCallThisFunctionSincePluginWasInjected() : \\\", \\n\\t\\t\\tplugins.sample().timeSpentUntilCallThisFunctionSincePluginWasInjected())\\n\\t\\treturn {status: 200, next: true, body: ctx.body, operation : ctx.operation.READ};\\n\\t}\", regex:(*regexp.Regexp)(nil)}"
>>>> GO JS PLUGIN >>>> ready :  true
>>>>> plugins.sample().ready :  true
>>>> GO JS PLUGIN >>>> ready :  true
>>>>> plugins.sample().timeSpentUntilCallThisFunctionSincePluginWasInjected() :  2000467481

```

<br/>
<a name="benchmark"/>

### 6. JS versus GO - information to help your choice

Very simple benchmark, just to compare the two types of filters.

JS code

``` javascript
{
	"pathPattern": ".*",
	"function" : function(ctx, plugins){
		for(var i = 0, i < 10000; i++){
			ctx.values.get('path').split("").join("");
			return {status: 200, next: true, body: ctx.body, operation : ctx.operation.READ};
		}
	}
}
```

GO code
``` go
package main

import (
	"strings"

	"github.com/kataras/iris"
)

// PluginModel PluginModel
type PluginModel struct {
	Next      bool
	Body      string
	Status    int
	Operation int
}

// Filter Filter
type Filter interface {
	Config() (Name string, Order int, PathPattern string, Invoke int)
	Exec(ctx iris.Context, requestBody string) (Next bool, Body string, Status int, Operation int, Err error)
}

// Exec Exec
func (filter PluginModel) Exec(ctx iris.Context, requestBody string) (Next bool, Body string, Status int, Operation int, Err error) {
	for i := 0; i < 10000; i++ {
		strings.Join(strings.Split(ctx.Values().GetString("path"), ""), "")
	}
	return true, requestBody, 200, 0, nil
}

// Config Config
func (filter PluginModel) Config() (Name string, Order int, PathPattern string, Invoke int) {
	return "GO_FILTER_BENCHMARK", 1, "/containers/json", 0
}

// Plugin exported as symbol
var Plugin PluginModel
```

No filters
``` terminal
[bruno@labbsr0x wrk2]$ wrk -t8 -c1000 -d30s -R10000 http://localhost:8080/v1.39/containers/json?all=1
Running 30s test @ http://localhost:8080/v1.39/containers/json?all=1
  8 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    13.38s     3.83s   20.92s    58.00%
    Req/Sec   397.50      2.92   403.00     75.00%
  93746 requests in 30.00s, 214.33MB read
Requests/sec:   3124.75
Transfer/sec:      7.14MB
```

JS results
``` terminal
[bruno@labbsr0x wrk2]$ wrk -t8 -c1000 -d30s -R10000 http://localhost:8080/v1.39/containers/json?all=1
Running 30s test @ http://localhost:8080/v1.39/containers/json?all=1
  8 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    16.51s     4.75s   25.33s    57.85%
    Req/Sec   187.38      1.11   189.00    100.00%
  44345 requests in 30.00s, 101.46MB read
Requests/sec:   1477.99
Transfer/sec:      3.38MB
```

GO results
``` terminal
[bruno@labbsr0x wrk2]$ wrk -t8 -c1000 -d30s -R10000 http://localhost:8080/v1.39/containers/json?all=1
Running 30s test @ http://localhost:8080/v1.39/containers/json?all=1
  8 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    19.00s     5.60s   29.18s    56.96%
    Req/Sec    22.38      0.70    23.00    100.00%
  5621 requests in 30.19s, 12.86MB read
  Socket errors: connect 0, read 0, write 0, timeout 8343
Requests/sec:    186.16
Transfer/sec:    436.12KB
```
Weird. Wasn't expecting this.

A simple test : the same benchmark code that was running inside the plugin but now running directly in go-horse code. No go plugins involved here.

```terminal
[bruno@labbsr0x wrk2]$ wrk -t8 -c100 -d5s -R10000 http://localhost:8080/v1.39/containers/json?all=1
Running 5s test @ http://localhost:8080/v1.39/containers/json?all=1
  8 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.87s     1.07s    3.82s    57.73%
    Req/Sec       -nan      -nan   0.00      0.00%
  12484 requests in 5.00s, 28.56MB read
Requests/sec:   2496.72
Transfer/sec:      5.71MB
```

Now the same JS filter but calling a go plugin's injected property and function   : 

``` terminal
[bruno@labbsr0x wrk2]$ wrk -t8 -c100 -d10s -R10000 http://localhost:8080/v1.39/containers/json?all=1
Running 10s test @ http://localhost:8080/v1.39/containers/json?all=1
  8 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.30s     2.48s    8.66s    57.81%
    Req/Sec       -nan      -nan   0.00      0.00%
  13887 requests in 10.00s, 31.77MB read
Requests/sec:   1388.36
Transfer/sec:      3.18MB
```